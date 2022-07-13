import sqlite3
import sys
import re

### Reading Data
# open db
db = sys.argv[1]
con = sqlite3.connect(db)
cur = con.cursor()

# read recipes
cur.execute(
    """
    select r.id, r.title, t.tag, i.ingredient
    from recipes r
    inner join ingredients i on r.id = i.recipeid
    left join tags t on r.id = t.recipeid
    """
)

rows = cur.fetchall()
cur.close()

### Data Cleanup
from nltk.corpus import stopwords
from nltk.tokenize import word_tokenize

# helpers for removing things
only_alpha = re.compile("[^a-zA-Z\\s]")

# https://en.wikipedia.org/wiki/Cooking_weights_and_measures
measures = [
    # measures
    "drop",
    "dr",
    "gt",
    "gtt",
    "smidgen" "smdg",
    "smi",
    "gtt",
    "ds",
    "pinch",
    "pn" "dash",
    'ds",' "saltspoon",
    "scruple",
    "ssp",
    "coffeespoon",
    "csp",
    "dram",
    "fluid",
    "fl" "dr",
    "teaspoon",
    "tsp",
    "t",
    "dessertspoon",
    "dsp",
    "dssp",
    "dstspon",
    "tablespoon",
    "tbsp",
    "ounce",
    "oz",
    "wineglass",
    "wgf",
    "gill",
    "teacup",
    "tcf",
    "cup",
    "c",
    "pint",
    "pt",
    "quart",
    "qt",
    "gallon",
    "gal",
]

special_instructions = [
    "firm",
    "firmly",
    "light",
    "lightly",
    "packed",
    "even",
    "level",
    "round",
    "rounded",
    "heap",
    "heaping",
    "heaped",
    "sifted",
    "sift",
    "scoop",
    "whole",
]

common_ingredients = ["salt", "pepper", "water", "oil"]

# pluralize
def pluralize(words):
    return [w + "s" for w in words] + [w + "es" for w in words]


special_instructions += pluralize(special_instructions)
measures += pluralize(measures)
common_ingredients += pluralize(common_ingredients)

# expand stop words here (e.g. measurments and common ingredients (minced, etc))
# look into collecting existing lists
stop_words = (
    set(stopwords.words("english"))
    | set(special_instructions)
    | set(measures)
    | set(common_ingredients)
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
                for w in word_tokenize(only_alpha.sub("", ingredient).lower())
                if not w in stop_words and len(w) > 2
            ]
        ),
    }

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

### Graph Generation
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


### Visualization
# https://melaniewalsh.github.io/Intro-Cultural-Analytics/06-Network-Analysis/00-Network-Analysis.html
from bokeh.transform import linear_cmap
from bokeh.io import output_file, save, show, curdoc
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

# title
output_file(sys.argv[3])
curdoc().theme = "light_minimal"
title = "Recipes Visualization"

# Create a plot — set dimensions, toolbar, and title
plot = figure(
    tools="pan,wheel_zoom,save,reset,hover",
    active_scroll="wheel_zoom",
    x_range=Range1d(-10.1, 10.1),
    y_range=Range1d(-10.1, 10.1),
    title=title,
)

# community based coloring
# Create empty dictionaries
modularity_class = {}
modularity_color = {}
# Loop through each community in the network
for community_number, community in enumerate(communities):
    # For each member of the community, add their community number and a distinct color
    for name in community:
        modularity_class[name] = community_number
        modularity_color[name] = Viridis8[community_number]

nx.set_node_attributes(g, modularity_class, "modularity_class")
nx.set_node_attributes(g, modularity_color, "modularity_color")

# Create a network graph object with spring layout
# https://networkx.github.io/documentation/networkx-1.9/reference/generated/networkx.drawing.layout.spring_layout.html
network_graph = from_networkx(g, nx.spring_layout, scale=10, center=(0, 0))

# settings
size_by = "NumKeyWords"

# color_by = "modularity_color"
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
    line_alpha=edge_opacity_by, line_color=edge_color_by, line_width=1
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
toolips = """
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
    </div>
    """


hover = plot.select_one(HoverTool)
hover.show_arrow = False
hover.tooltips = toolips

# layout
column(plot, sizing_mode="stretch_both")

# show(plot)
save(plot)
