# %% [markdown]
# # Read Args
# Setup args used by program.

# %%
# read args
import sys
import math
import time
from datetime import datetime

# db = sys.argv[1]
db = "../scrape/seriouseats/seriouseats.db"

limit = 100

# num_clusters = int(sys.argv[2])
num_clusters = 12

# 0 means keep all edges
min_edge_weight = 0

# progress bars
def progress(count, total, status=''):
    width = 60
    percents = round(100.0 * (count + 1) / float(total), 1)
    progress = (count + 1) / total
    progress = min(1, max(0, progress))
    whole_width = math.floor(progress * width)
    remainder_width = (progress * width) % 1
    part_width = math.floor(remainder_width * 8)
    part_char = [" ", "▏", "▎", "▍", "▌", "▋", "▊", "▉"][part_width]
    if (width - whole_width - 1) < 0:
        part_char = ""
    bar = "[" + "█" * whole_width + part_char + " " * (width - whole_width - 1) + "]"

    sys.stdout.write('%s %s%s %s\r' % (bar, percents, '%', status))
    if (width - whole_width - 1) < 0:
        sys.stdout.write("\n")
    sys.stdout.flush()

# filename with timestamp
def get_filename(filename, ext):
    date = datetime.now().strftime("%Y_%m_%d-%I-%M-%S_%p")
    return f"{filename}_{date}.{ext}"


# %% [markdown]
# # Read Data
# Read in data from sqlite3 database

# %%
%time
### Read Data
import sqlite3

# open db
con = sqlite3.connect(db)
cur = con.cursor()

# read recipes
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
print("Loaded {:d} rows".format(len(rows)))

# %% [markdown]
# # Cleanup Data
# Current process removes nltk stopwords, custom stopwords, and uses nltk to strip non-nouns.

# %%
%time
### Data Cleanup
import re
import os

from nltk import pos_tag
from nltk.corpus import stopwords
from nltk.tokenize import word_tokenize
from pathlib import Path

# helpers for removing things
only_alpha = re.compile("[^a-zA-Z\\s]")

# pluralize
def pluralize(words):
    return [w + "s" for w in words] + [w + "es" for w in words]

custom_stopwords_filepath = os.path.join(
    Path().resolve(), "custom_stopwords.txt"
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
for i, [id, title, tag, ingredient] in enumerate(rows):
    progress(i, len(rows), "Cleaning")
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

# %% [markdown]
# # Clustering
# Perform clustering via [kmeans](https://scikit-learn.org/stable/modules/generated/sklearn.cluster.KMeans.html)

# %%
%time
### Clustering
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
    progress(i, len(results), "Encoding")
    for word in result["values"]:
        if word in words_to_index:
            encoded[i][words_to_index[word]] = 1

# cluster
kmeans = KMeans(n_clusters=num_clusters, random_state=0).fit(encoded)
clusters = [
    (pair[0], pair[1][0], pair[1][1]["title"], pair[1][1]["values"])
    for pair in zip(kmeans.labels_, results.items())
]

# %% [markdown]
# # Graph Generation
# Generate graph using [networkx](https://networkx.org/documentation/stable/index.html)

# %%
%time
### Graph Generation
import networkx as nx
from networkx.algorithms import community

g = nx.Graph()
edges = {}
node_labels = []
for i, (cluster, id, title, values) in enumerate(clusters):
    progress(i, len(results), "Generating Nodes and Edges")
    label = (title[:20] + "..") if len(title) > 20 else title
    g.add_node(
        i,
        Title=title,
        Label=label,
        ID=id,
        Cluster=cluster,
        KeyWords=", ".join(values),
        NumKeyWords=len(values),
    )

    node_labels.append(label)

    for word in values:
        for j, row in enumerate(encoded):
            if i != j and row[words_to_index[word]] == 1:
                if (j, i) in edges:
                    edges[(j, i)] += 1
                elif (i, j) in edges:
                    edges[(i, j)] += 1
                else:
                    edges[(i, j)] = 1

# weight by matches
# max_weight = max(edges.items(), key=operator.itemgetter(1))[1]
# edges = {(i, j): w / max_weight for (i, j), w in edges.items()}

# weight by cluster
#edges = {
#    (i, j): int(clusters[i][0] == clusters[j][0])
#    for (i, j), _ in edges.items()
#}

# weight by mixture
max_weight = max(edges.items(), key=operator.itemgetter(1))[1]
edges = {
    (i, j): w / max_weight * (int(clusters[i][0] == clusters[j][0]) + 1)
    for (i, j), w in edges.items()
}

for k, ((i, j), w) in enumerate(edges.items()):
    progress(k, len(edges), "Attaching Edges")
    if w >= min_edge_weight:
        g.add_edge(i, j, weight=w)

# calc communities
communities = community.greedy_modularity_communities(g)
print("{:d} nodes and {:d} edges".format(len(g.nodes()), len(g.edges())))
print("Retained {:d} of {:d} possible edges".format(len(g.edges()), len(edges)))

# %% [markdown]
# # Bokeh Render
# Render using [bokeh](https://bokeh.org).
# 
# Learned from this nice site https://melaniewalsh.github.io/Intro-Cultural-Analytics/06-Network-Analysis/00-Network-Analysis.html.
# 
# Works okay for low number of nodes/edges.

# %%
### Bokeh Visualization
## %%script false --no-raise-error
%time
from bokeh.transform import linear_cmap
from bokeh.io import output_notebook, output_file, save, show, curdoc
from bokeh.themes import built_in_themes
from bokeh.models import (
    Range1d,
    Circle,
    HoverTool,
    MultiLine,
    LabelSet,
    ColumnDataSource,
    NodesAndLinkedEdges,
)
from bokeh.plotting import figure, column
from bokeh.plotting import from_networkx
from bokeh.palettes import Viridis8

output_notebook()

# title
curdoc().theme = "dark_minimal"
title = "Recipes Visualization"

# Create a plot — set dimensions, toolbar, and title
plot = figure(
    tools="pan,wheel_zoom,save,reset,hover",
    active_scroll="wheel_zoom",
    x_range=Range1d(-10.1, 10.1),
    y_range=Range1d(-10.1, 10.1),
    title=title,
)

# Create a network graph object with spring layout
# https://networkx.github.io/documentation/networkx-1.9/reference/generated/networkx.drawing.layout.spring_layout.html
network_graph = from_networkx(
    g,
    nx.spring_layout,
    scale=10,
    center=(0, 0),
)

# settings
size_by = "NumKeyWords"

color_by = "modularity_color"
node_color_by = linear_cmap("Cluster", Viridis8, 0, num_clusters)

edge_color_by = "black"
edge_opacity_by = "weight"

node_highlight_color = "white"
edge_highlight_color = "black"

# Set node size and color
network_graph.node_renderer.glyph = Circle(
    size=size_by,
    fill_color=node_color_by,
)

# Set node highlight colors
network_graph.node_renderer.hover_glyph = Circle(
    size=size_by, fill_color=node_highlight_color, line_width=2
)
network_graph.node_renderer.selection_glyph = Circle(
    size=size_by, fill_color=node_highlight_color, line_width=2
)

# Set edge opacity and width
network_graph.edge_renderer.glyph = MultiLine(
    line_alpha=0.1, line_color=edge_color_by, line_width=1
)

# Set edge highlight colors
network_graph.edge_renderer.selection_glyph = MultiLine(
    line_color=edge_highlight_color, line_width=2
)
network_graph.edge_renderer.hover_glyph = MultiLine(
    line_color=edge_highlight_color, line_width=2
)

# Highlight nodes and edges
network_graph.selection_policy = NodesAndLinkedEdges()
network_graph.inspection_policy = NodesAndLinkedEdges()

# Add network graph to the plot
plot.renderers.append(network_graph)

x, y = zip(*network_graph.layout_provider.graph_layout.values())
source = ColumnDataSource(
    {"x": x, "y": y, "name": [node_labels[i] for i in range(len(x))]}
)

labels = LabelSet(
    x="x",
    y="y",
    text="name",
    source=source,
    background_fill_color="white",
    text_font_size="10px",
    background_fill_alpha=0.7,
    text_align="center",
)
plot.renderers.append(labels)

# setup tooltip
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
    <img src="http://localhost/images/recipe/@ID" />
    </div>
    """

hover = plot.select_one(HoverTool)
hover.show_arrow = False
hover.tooltips = tooltips

# layout
column(plot)

# display
show(plot)

# %% [markdown]
# # Datashader Rendering
# Render using [datashader](https://datashader.org)

# %%
### Datashader Render
import datashader as ds
import datashader.transfer_functions as tf
from datashader.layout import random_layout, circular_layout, forceatlas2_layout
from datashader.bundling import connect_edges, hammer_bundle
import pandas as pd
from holoviews.plotting.util import process_cmap

cvsopts = dict(plot_height=800, plot_width=800)


def nodesplot(nodes, name=None, canvas=None, cat=None):
    canvas = ds.Canvas(**cvsopts) if canvas is None else canvas
    aggregator = None if cat is None else ds.count_cat(cat)
    agg = canvas.points(nodes, "x", "y", aggregator)

    color_key = ["#00FF00"]

    if cat is not None:
        cats = set([c for c in nodes[cat]])
        color_key = process_cmap("Viridis", provider="bokeh", ncolors=len(cats))

    return tf.spread(
        tf.shade(agg, color_key=color_key),
        px=3, name=name,
    )


def edgesplot(edges, name=None, canvas=None):
    canvas = ds.Canvas(**cvsopts) if canvas is None else canvas
    return tf.shade(
        canvas.line(edges, "x", "y", agg=ds.count()),
        name=name,
        cmap=process_cmap("Viridis", provider="bokeh"),
    )


def graphplot(nodes, edges, name="", canvas=None, cat=None):
    if canvas is None:
        xr = nodes.x.min(), nodes.x.max()
        yr = nodes.y.min(), nodes.y.max()
        canvas = ds.Canvas(x_range=xr, y_range=yr, **cvsopts)

    np = nodesplot(nodes, name + " nodes", canvas, cat)
    ep = edgesplot(edges, name + " edges", canvas)
    return tf.stack(ep, np, how="over", name=name)


def nx_layout(graph, cat=None):
    layout = nx.spring_layout(graph, iterations=100)
    data = [
        [node[0]]  # index
        + layout[node[0]].tolist()  # x,y
        + ([] if cat is None else [node[1][cat]])
        for node in graph.nodes(data=True)
    ]

    columns = ["id", "x", "y"] + ([] if cat is None else [cat])
    nodes = pd.DataFrame(data, columns=columns)
    nodes.set_index("id", inplace=True)

    if cat is not None:
        nodes[cat] = nodes[cat].astype("category")
 
    edges = pd.DataFrame(list(graph.edges), columns=["source", "target"])
    return nodes, edges


def nx_plot(graph, name="", cat=None):
    nodes, edges = nx_layout(graph, cat)

    # direct = connect_edges(nodes, edges)
    # bundled_bw005 = hammer_bundle(nodes, edges)
    bundled_bw030 = hammer_bundle(nodes, edges, initial_bandwidth=0.30)

    return (graphplot(nodes, bundled_bw030, name, cat=cat), nodes, edges)


(plot, nodes, edges) = nx_plot(g, "Recipes", cat="Cluster")


# %% [markdown]
# # Datashader Image Output
# Output datashader results to png

# %%
### Datashader Visualization
from PIL import Image, ImageFont, ImageDraw


def draw_text(img, text, x, y, size=30, face="Regular", color=(230, 230, 230)):
    """Helper to draw text labels using PIL."""
    # This font path assumes you're using a Mac with Open Sans installed.
    fnt = ImageFont.load_default()
    try:
        fnt = ImageFont.truetype(f"~/Library/Fonts/OpenSans-{face}.ttf", size)
    except:
        pass
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


imgs = [draw_text(tf.set_background(plot, (26, 24, 38)).to_pil(), plot.name, 15, 0)]
output = tile_images(imgs)
output.save(get_filename("output/cluster", "png"))
output


# %% [markdown]
# # Datashader Interactive Output
# Make datashader plots interactive with bokeh

# %%
import holoviews as hv
from holoviews import opts
from holoviews.plotting.util import process_cmap
from holoviews.operation.datashader import datashade, bundle_graph
from bokeh.models import HoverTool
hv.extension("bokeh")
hv.renderer('bokeh').theme = 'dark_minimal'

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

# defaults
kwargs = dict(width=800,height=800, xaxis=None, yaxis=None)
nodes_kwargs = {**kwargs, **dict(size = 15)}
opts.defaults(
    opts.Nodes(**nodes_kwargs), opts.Graph(**kwargs)
)

# settings
node_cmap=process_cmap("Viridis", provider="bokeh")

# create graph
graph = hv.Graph.from_networkx(g, nx.layout.fruchterman_reingold_layout)
graph.opts(
    width=800,height=800,
    cmap=node_cmap,
    edge_line_width=1, 
    node_color='Cluster',
    tools=[hover],
    active_tools=["wheel_zoom"],
    bgcolor=(26, 24, 38)
)


graph = bundle_graph(graph)
labels = hv.Labels(graph.nodes, ['x', 'y'], 'Label')

# combine graphs
graph = (
    (datashade(graph, normalization='linear', cmap=node_cmap, width=800, height=800) * graph.nodes).opts(
        opts.Nodes(color='Cluster', size='NumKeyWords', cmap=node_cmap, active_tools=["wheel_zoom"], tools=[hover]),
    )
    #* graph.select(circle='Cluster').opts(
    #    edge_alpha=0, edge_hover_alpha=1,
    #    edge_hover_line_color="white",
    #    node_color='Cluster'
    #)
    * labels.opts(text_font_size='8pt', text_color='grey')
)


hv.save(graph, get_filename("output/cluster", "html"))
graph


