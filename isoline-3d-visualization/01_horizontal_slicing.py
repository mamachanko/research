"""
Approach 1: Horizontal Slicing — Classic Topographic Contours

Slices a 3D surface at regular z-levels to produce a contour map.
Demonstrates the technique on multiple surfaces: a peak landscape,
a torus cross-section, and the Himmelblau function.
"""

import numpy as np
import matplotlib.pyplot as plt
from matplotlib import cm
import matplotlib.colors as mcolors


def make_peak_landscape(n=400):
    x = np.linspace(-3, 3, n)
    y = np.linspace(-3, 3, n)
    X, Y = np.meshgrid(x, y)
    Z = (
        3 * (1 - X) ** 2 * np.exp(-(X**2) - (Y + 1) ** 2)
        - 10 * (X / 5 - X**3 - Y**5) * np.exp(-(X**2) - Y**2)
        - 1 / 3 * np.exp(-((X + 1) ** 2) - Y**2)
    )
    return X, Y, Z


def make_himmelblau(n=400):
    x = np.linspace(-5, 5, n)
    y = np.linspace(-5, 5, n)
    X, Y = np.meshgrid(x, y)
    Z = (X**2 + Y - 11) ** 2 + (X + Y**2 - 7) ** 2
    return X, Y, Z


def make_saddle(n=400):
    x = np.linspace(-2, 2, n)
    y = np.linspace(-2, 2, n)
    X, Y = np.meshgrid(x, y)
    Z = X**2 - Y**2
    return X, Y, Z


fig, axes = plt.subplots(1, 3, figsize=(18, 6))
fig.patch.set_facecolor("#0a0a0a")

surfaces = [
    ("Peak Landscape", make_peak_landscape, 25, "cividis"),
    ("Himmelblau Function", make_himmelblau, 30, "inferno"),
    ("Hyperbolic Paraboloid", make_saddle, 20, "viridis"),
]

for ax, (title, fn, n_levels, cmap_name) in zip(axes, surfaces):
    X, Y, Z = fn()
    ax.set_facecolor("#0a0a0a")
    ax.set_aspect("equal")

    cmap = plt.get_cmap(cmap_name)
    levels = np.linspace(Z.min(), Z.max(), n_levels)

    # Filled contours for subtle shading
    ax.contourf(X, Y, Z, levels=levels, cmap=cmap, alpha=0.15)

    # Isoline contours
    cs = ax.contour(X, Y, Z, levels=levels, cmap=cmap, linewidths=0.8)

    ax.set_title(title, color="white", fontsize=14, fontweight="bold", pad=12)
    ax.tick_params(colors="white", labelsize=8)
    for spine in ax.spines.values():
        spine.set_color("#333")

fig.suptitle(
    "Horizontal Slicing — Z-Level Contours",
    color="white",
    fontsize=18,
    fontweight="bold",
    y=0.98,
)
plt.tight_layout(rect=[0, 0, 1, 0.93])
plt.savefig(
    "/home/user/research/isoline-3d-visualization/01_horizontal_slicing.png",
    dpi=150,
    bbox_inches="tight",
    facecolor=fig.get_facecolor(),
)
plt.close()
print("Saved 01_horizontal_slicing.png")
