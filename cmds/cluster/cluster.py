import sqlite3
import sys
import re
import os
import time

### Reading Data
start = time.time()
# open db
db = sys.argv[1]
con = sqlite3.connect(db)
cur = con.cursor()

# read recipes
limit = 50
cur.execute(
    """
    select r.id, r.title, t.tag, i.ingredient
    from (
        select *
        from recipes
        limit ?
    ) r
    inner join ingredients i on r.id = i.recipeid
    left join tags t on r.id = t.recipeid
    """,
    [limit],
)

rows = cur.fetchall()
cur.close()
print("Data Load: ", time.time() - start)

### Data Cleanup
start = time.time()
from nltk import pos_tag
from nltk.corpus import stopwords
from nltk.tokenize import word_tokenize

# helpers for removing things
only_alpha = re.compile("[^a-zA-Z\\s]")

# pluralize
def pluralize(words):
    return [w + "s" for w in words] + [w + "es" for w in words]


# special_instructions += pluralize(special_instructions)
# measures += pluralize(measures)
# common_ingredients += pluralize(common_ingredients)

custom_stopwords_filepath = os.path.join(
    os.path.dirname(__file__), "custom_stopwords.txt"
)

custom_stopwords = set(
    [
        w
        for w in open(custom_stopwords_filepath).read().split("\n")
        if not w.startswith("#") and not w.isspace() and not w == ""
    ]
)

# expand stop words here (e.g. measurments and common ingredients (minced, etc))
# look into collecting existing lists
stop_words = (
    set(stopwords.words("english"))
    | set(custom_stopwords)
    | set(pluralize(custom_stopwords))
)

# create results to work with
results = {}
for [id, title, tag, ingredient] in rows:
    results[id] = {
        "title": title,
        "values":
        # existing results
        (results[id].get("values") if id in results else set())
        # tags
        # | set([tag.lower()])
        # ingredients
        | set(
            [
                w
                for w, pos in pos_tag(
                    word_tokenize(only_alpha.sub(" ", ingredient).lower())
                )
                if not w in stop_words
                and len(w) > 2
                and (pos == "NN" or pos == "NNP" or pos == "NNS" or pos == "NNPS")
            ]
        ),
    }

print("Cleanup: ", time.time() - start)

### Clustering
start = time.time()
from sklearn.cluster import KMeans
import numpy as np
import operator

# encode data
words = sorted(
    set([value for result in results.values() for value in result["values"]])
)

words_to_index = {key: i for i, key in enumerate(words)}
encoded = np.zeros((len(results), len(words)))
for i, (_, result) in enumerate(results.items()):
    for word in result["values"]:
        if word in words_to_index:
            encoded[i][words_to_index[word]] = 1

# cluster
num_clusters = int(sys.argv[2])
kmeans = KMeans(n_clusters=num_clusters, random_state=0).fit(encoded)
clusters = [
    (pair[0], pair[1][0], pair[1][1]["title"], pair[1][1]["values"])
    for pair in zip(kmeans.labels_, results.items())
]

print("Clustering: ", time.time() - start)

### Graph Generation
start = time.time()
import networkx as nx
from networkx.algorithms import community

g = nx.Graph()
edges = {}
node_labels = []
for i, (cluster, id, title, values) in enumerate(clusters):
    g.add_node(
        i,
        Label=title,
        ID=id,
        Title=title,
        Cluster=cluster,
        KeyWords=", ".join(values),
        NumKeyWords=len(values),
    )

    node_labels.append((title[:20] + "..") if len(title) > 20 else title)

    for word in values:
        for j, row in enumerate(encoded):
            if i != j and row[words_to_index[word]] == 1:
                if (j, i) in edges:
                    edges[(j, i)] += 1
                elif (i, j) in edges:
                    edges[(i, j)] += 1
                else:
                    edges[(i, j)] = 1

max_weight = max(edges.items(), key=operator.itemgetter(1))[1]
edges = {(i, j): w / max_weight for (i, j), w in edges.items()}

for (i, j), w in edges.items():
    g.add_edge(i, j, weight=w)

# calc communities
communities = community.greedy_modularity_communities(g)

print("Graph Generation: ", time.time() - start)

### Visualization
start = time.time()
from PIL import Image, ImageFont, ImageDraw


def draw_text(img, text, x, y, size=30, face="Regular", color=(230, 230, 230)):
    """Helper to draw text labels using PIL."""
    # This font path assumes you're using a Mac with Open Sans installed.
    fnt = ImageFont.truetype(f"~/Library/Fonts/OpenSans-{face}.ttf", size)
    d = ImageDraw.Draw(img)
    bbox = d.textbbox((x, y), text, fnt)
    d.text((bbox[0], bbox[1]), text, font=fnt, fill=color)
    return img


def tile_images(images):
    # Tile the three images side by side
    widths, heights = zip(*(i.size for i in images))
    width = sum(widths)
    height = max(heights)
    output = Image.new("RGB", (width, height), (26, 24, 38))
    x_offset = 0
    for im in images:
        output.paste(im, (x_offset, 0))
        x_offset += im.size[0]

    return output


### Datashader
import datashader as ds
import datashader.transfer_functions as tf
from datashader.layout import random_layout, circular_layout, forceatlas2_layout
from datashader.bundling import connect_edges, hammer_bundle
import pandas as pd
from itertools import chain

cvsopts = dict(plot_height=800, plot_width=800)


def nodesplot(nodes, name=None, canvas=None, cat=None):
    canvas = ds.Canvas(**cvsopts) if canvas is None else canvas
    aggregator = None if cat is None else ds.count_cat(cat)
    agg = canvas.points(nodes, "x", "y", aggregator)
    return tf.spread(tf.shade(agg, cmap=["#00FF00"]), px=3, name=name)


def edgesplot(edges, name=None, canvas=None):
    canvas = ds.Canvas(**cvsopts) if canvas is None else canvas
    return tf.shade(canvas.line(edges, "x", "y", agg=ds.count()), name=name)


def graphplot(nodes, edges, name="", canvas=None, cat=None):
    if canvas is None:
        xr = nodes.x.min(), nodes.x.max()
        yr = nodes.y.min(), nodes.y.max()
        canvas = ds.Canvas(x_range=xr, y_range=yr, **cvsopts)

    np = nodesplot(nodes, name + " nodes", canvas, cat)
    ep = edgesplot(edges, name + " edges", canvas)
    return tf.stack(ep, np, how="over", name=name)


def nx_layout(graph):
    layout = nx.spring_layout(graph)
    data = [[node] + layout[node].tolist() for node in graph.nodes]

    nodes = pd.DataFrame(data, columns=["id", "x", "y"])
    nodes.set_index("id", inplace=True)

    edges = pd.DataFrame(list(graph.edges), columns=["source", "target"])
    return nodes, edges


def nx_plot(graph, name=""):
    # print(graph.name, len(graph.edges))
    nodes, edges = nx_layout(graph)

    direct = connect_edges(nodes, edges)
    bundled_bw005 = hammer_bundle(nodes, edges)
    bundled_bw030 = hammer_bundle(nodes, edges, initial_bandwidth=0.30)

    return [
        graphplot(nodes, direct, graph.name),
        graphplot(nodes, bundled_bw005, "Bundled bw=0.05"),
        graphplot(nodes, bundled_bw030, "Bundled bw=0.30"),
    ]


plots = nx_plot(g)

imgs = [
    draw_text(tf.set_background(plot, (26, 24, 38)).to_pil(), plot.name, 15, 0)
    for plot in plots
]
output = tile_images(imgs)
output.save("tmp.png")


from pixcat import Image

Image("tmp.png").show()

print("Rendering: ", time.time() - start)

import holoviews as hv
from holoviews import opts
from holoviews.plotting.util import process_cmap
from holoviews.operation.datashader import datashade, bundle_graph
from bokeh.models import HoverTool

hv.extension("bokeh")
hv.renderer("bokeh").theme = "dark_minimal"

tooltips = """
    <div style="font-family: Helvetica;
                border: 1px solid black;
                background-color: white;
                width : 200px;
                position: fixed;
                left: 65px;
                top: 45px;
                padding: 10px">

    <span style="font-size: 16px;"><b>Title:</b> @Title</span><br>
    <span style="font-size: 16px;"><b>Cluster:</b> @Cluster</span><br>
    <span style="font-size: 16px;"><b>Keywords:</b> @KeyWords</span><br>
    <img  style="max-width: 200px;" src="http://localhost/images/recipe/@ID" />
    </div>
    """

hover = HoverTool(tooltips=tooltips)
kwargs = dict(width=800, height=800, xaxis=None, yaxis=None)
nodes_kwargs = {**kwargs, **dict(size=15)}
opts.defaults(opts.Nodes(**nodes_kwargs), opts.Graph(**kwargs))

graph = hv.Graph.from_networkx(g, nx.layout.fruchterman_reingold_layout)
graph.opts(
    cmap=process_cmap("Viridis", provider="bokeh"),
    width=800,
    height=800,
    edge_line_width=1,
    node_color="Cluster",
    tools=[hover],
    active_tools=["wheel_zoom"],
)

graph = bundle_graph(graph)
labels = hv.Labels(graph.nodes, ["x", "y"], "Label")
graph = (
    (
        datashade(graph, normalization="linear", width=800, height=800) * graph.nodes
    ).opts(opts.Nodes(color="Cluster", cmap=colors))
    * graph.opts(edge_alpha=0).select(circle="Cluster").opts(node_color="Cluster")
    * labels.opts(text_font_size="8pt", text_color="grey")
)

hv.save(graph, "out.html")
