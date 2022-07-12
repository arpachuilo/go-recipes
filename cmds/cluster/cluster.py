from sklearn.cluster import KMeans
from nltk.corpus import stopwords
from nltk.tokenize import word_tokenize
import numpy as np
import sqlite3
import sys
import re
from pprint import pprint
import networkx as nx
import matplotlib.pyplot as plt
import operator
from bokeh.models import Range1d, Circle, ColumnDataSource, MultiLine
from bokeh.plotting import figure
from bokeh.plotting import from_networkx

# open db
db = sys.argv[1]
con = sqlite3.connect(db)
cur = con.cursor()

# read recipes
cur.execute(
    """
    select r.id, r.title, t.tag, i.ingredient
    from recipes r
    inner join tags t on r.id = t.recipeid
    inner join ingredients i on r.id = i.recipeid
    """
)

rows = cur.fetchall()

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

# pprint(results)

# encoded data
words = sorted(
    set([value for result in results.values() for value in result["values"]])
)
# print(words)

words_to_index = {key: i for i, key in enumerate(words)}
encoded = np.zeros((len(results), len(words)))
for i, (_, result) in enumerate(results.items()):
    for word in result["values"]:
        if word in words_to_index:
            encoded[i][words_to_index[word]] = 1

# cluster
kmeans = KMeans(n_clusters=int(sys.argv[2]), random_state=0).fit(encoded)
clusters = [
    (pair[0], pair[1][0], pair[1][1]["title"], pair[1][1]["values"])
    for pair in zip(kmeans.labels_, results.items())
]

# pprint(clusters)

# viz
g = nx.Graph()

labels = {}
edges = {}
for i, (cluster, id, title, values) in enumerate(clusters):
    g.add_node(i, Label=title)
    labels[i] = title
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
print(max_weight)
edges = {(i, j): w / max_weight for (i, j), w in edges.items()}

for (i, j), w in edges.items():
    g.add_edge(i, j, weight=w)

pos = nx.spring_layout(g)

# draw edges
elarge = [(u, v) for (u, v, d) in g.edges(data=True) if d["weight"] > 0.75]
esmall = [(u, v) for (u, v, d) in g.edges(data=True) if d["weight"] <= 0.75]

nx.draw_networkx_edges(
    g, pos, edgelist=esmall, width=1, alpha=0.5, edge_color="b", style="dashed"
)
nx.draw_networkx_edges(g, pos, edgelist=elarge, width=6)

# draw nodes
nx.draw_networkx_nodes(
    g, pos=pos, node_size=800, node_color=kmeans.labels_, cmap=plt.viridis()
)

# node labels
labels_params = {
    "font_family": "sans-serif",
    "font_weight": 800,
    "font_color": "white",
    "font_size": 16,
}

nx.draw_networkx_labels(g, pos=pos, labels=labels)

# edge labels
edge_labels = {
    k: "{:.2f}".format(v) for k, v in nx.get_edge_attributes(g, "weight").items()
}
nx.draw_networkx_edge_labels(g, pos, edge_labels)

# plt.show()

# title
title = "Recipes Visualization"

# Establish which categories will appear when hovering over each node
HOVER_TOOLTIPS = [("Recipe", "@index")]

# Create a plot — set dimensions, toolbar, and title
plot = figure(
    tooltips=HOVER_TOOLTIPS,
    tools="pan,wheel_zoom,save,reset",
    active_scroll="wheel_zoom",
    x_range=Range1d(-10.1, 10.1),
    y_range=Range1d(-10.1, 10.1),
    title=title,
)
