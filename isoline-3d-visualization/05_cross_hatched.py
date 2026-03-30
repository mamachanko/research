"""
Approach 5: Cross-Hatched Multi-Axis Isolines

Overlays contour sets computed from projections along different axes
(x, y, and diagonal), creating a cross-hatched effect that reveals
the 3D shape through line density.
"""

import numpy as np
import matplotlib.pyplot as plt


def make_surface(n=400):
    x = np.linspace(-3, 3, n)
    y = np.linspace(-3, 3, n)
    X, Y = np.meshgrid(x, y)
    Z = (
        2 * np.exp(-((X - 0.5) ** 2 + Y**2) / 1.2)
        + 1.5 * np.exp(-((X + 1.2) ** 2 + (Y - 1) ** 2) / 0.8)
        - 0.5 * np.exp(-((X - 1.5) ** 2 + (Y + 1.5) ** 2) / 0.5)
    )
    return X, Y, Z


def make_egg_carton(n=400):
    x = np.linspace(-3, 3, n)
    y = np.linspace(-3, 3, n)
    X, Y = np.meshgrid(x, y)
    Z = np.sin(X * 2) * np.cos(Y * 2)
    return X, Y, Z


fig, axes = plt.subplots(1, 2, figsize=(16, 8))
fig.patch.set_facecolor("#0a0a0a")

datasets = [
    ("Gaussian Landscape", make_surface),
    ("Egg Carton (sin*cos)", make_egg_carton),
]

axis_configs = [
    # (transform, color, label)
    (lambda X, Y, Z: Z, "#ff4466", "Z contours"),
    (lambda X, Y, Z: Z + X * 0.8, "#44aaff", "Z+X contours"),
    (lambda X, Y, Z: Z + Y * 0.8, "#44ff88", "Z+Y contours"),
]

for ax, (title, fn) in zip(axes, datasets):
    X, Y, Z = fn()
    ax.set_facecolor("#0a0a0a")
    ax.set_aspect("equal")

    for transform, color, label in axis_configs:
        field = transform(X, Y, Z)
        levels = np.linspace(field.min(), field.max(), 16)
        ax.contour(
            X, Y, field, levels=levels,
            colors=[color], linewidths=0.5, alpha=0.6,
        )

    ax.set_title(title, color="white", fontsize=14, fontweight="bold", pad=12)
    ax.tick_params(colors="white", labelsize=7)
    for spine in ax.spines.values():
        spine.set_color("#222")

    # Legend
    from matplotlib.lines import Line2D
    legend_elements = [
        Line2D([0], [0], color=c, linewidth=1, label=l)
        for _, c, l in axis_configs
    ]
    ax.legend(handles=legend_elements, loc="lower right", fontsize=8,
              facecolor="#1a1a1a", edgecolor="#333", labelcolor="white")

fig.suptitle(
    "Cross-Hatched Multi-Axis Isolines",
    color="white", fontsize=18, fontweight="bold", y=0.98,
)
plt.tight_layout(rect=[0, 0, 1, 0.93])
plt.savefig(
    "/home/user/research/isoline-3d-visualization/05_cross_hatched.png",
    dpi=150, bbox_inches="tight", facecolor=fig.get_facecolor(),
)
plt.close()
print("Saved 05_cross_hatched.png")
