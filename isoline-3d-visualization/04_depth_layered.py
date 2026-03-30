"""
Approach 4: Depth-Layered Contours

Contour lines where line thickness and opacity encode the z-level,
creating a sense of depth. Higher (closer) contours are thicker and
more opaque; lower (further) contours are thin and faint.
"""

import numpy as np
import matplotlib.pyplot as plt
from matplotlib.colors import to_rgba


def make_double_peak(n=500):
    x = np.linspace(-3, 3, n)
    y = np.linspace(-3, 3, n)
    X, Y = np.meshgrid(x, y)
    Z = (
        2.5 * np.exp(-((X - 1) ** 2 + Y**2))
        + 2.0 * np.exp(-((X + 1) ** 2 + (Y - 0.5) ** 2) / 0.6)
        + 1.0 * np.exp(-((X + 0.5) ** 2 + (Y + 1.5) ** 2) / 0.4)
    )
    return X, Y, Z


def make_ripple(n=500):
    x = np.linspace(-4, 4, n)
    y = np.linspace(-4, 4, n)
    X, Y = np.meshgrid(x, y)
    R = np.sqrt(X**2 + Y**2)
    Z = np.cos(R * 2) * np.exp(-R / 3)
    return X, Y, Z


def make_spiral_peak(n=500):
    x = np.linspace(-3, 3, n)
    y = np.linspace(-3, 3, n)
    X, Y = np.meshgrid(x, y)
    R = np.sqrt(X**2 + Y**2) + 1e-6
    theta = np.arctan2(Y, X)
    Z = np.exp(-R / 2) * np.cos(R * 3 - theta * 2) + np.exp(-R)
    return X, Y, Z


fig, axes = plt.subplots(1, 3, figsize=(18, 6))
fig.patch.set_facecolor("#0a0a0a")

surfaces = [
    ("Double Peak — Depth Layers", make_double_peak, "#00ddff", 20),
    ("Ripple — Depth Layers", make_ripple, "#ffaa00", 18),
    ("Spiral Peak — Depth Layers", make_spiral_peak, "#88ff44", 22),
]

for ax, (title, fn, base_color, n_levels) in zip(axes, surfaces):
    X, Y, Z = fn()
    ax.set_facecolor("#0a0a0a")
    ax.set_aspect("equal")

    levels = np.linspace(Z.min(), Z.max(), n_levels)
    z_range = Z.max() - Z.min()

    for i, level in enumerate(levels):
        t = (level - Z.min()) / z_range  # 0 = deep, 1 = high
        alpha = 0.15 + 0.85 * t
        linewidth = 0.3 + 2.0 * t
        rgba = to_rgba(base_color, alpha=alpha)

        ax.contour(
            X, Y, Z, levels=[level],
            colors=[rgba], linewidths=linewidth,
        )

    ax.set_title(title, color="white", fontsize=13, fontweight="bold", pad=12)
    ax.tick_params(colors="white", labelsize=7)
    for spine in ax.spines.values():
        spine.set_color("#222")

fig.suptitle(
    "Depth-Layered Contours — Thickness & Opacity Encode Height",
    color="white", fontsize=17, fontweight="bold", y=0.98,
)
plt.tight_layout(rect=[0, 0, 1, 0.93])
plt.savefig(
    "/home/user/research/isoline-3d-visualization/04_depth_layered.png",
    dpi=150, bbox_inches="tight", facecolor=fig.get_facecolor(),
)
plt.close()
print("Saved 04_depth_layered.png")
