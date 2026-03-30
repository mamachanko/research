"""
Approach 2: Joy Division / Ridgeline Plot

Cross-section profiles stacked vertically with occlusion, creating
a 3D illusion from pure line work. Inspired by the Unknown Pleasures
album cover and ridgeline plots.
"""

import numpy as np
import matplotlib.pyplot as plt


def make_terrain(n=300):
    x = np.linspace(-3, 3, n)
    y = np.linspace(-3, 3, 60)
    X, Y = np.meshgrid(x, y)
    Z = (
        2.0 * np.exp(-((X - 0.5) ** 2 + (Y - 0.3) ** 2) / 0.8)
        + 1.5 * np.exp(-((X + 1.0) ** 2 + (Y + 0.5) ** 2) / 0.5)
        + 0.8 * np.exp(-((X - 1.5) ** 2 + (Y + 1.0) ** 2) / 0.6)
        + 0.5 * np.sin(X * 3) * np.cos(Y * 2) * 0.3
    )
    return X, Y, Z


def make_wave(n=300):
    x = np.linspace(-4, 4, n)
    y = np.linspace(-4, 4, 60)
    X, Y = np.meshgrid(x, y)
    R = np.sqrt(X**2 + Y**2) + 1e-6
    Z = np.sin(R * 2.5) / R * 3
    return X, Y, Z


fig, axes = plt.subplots(1, 2, figsize=(16, 9))
fig.patch.set_facecolor("#0a0a0a")

datasets = [
    ("Gaussian Peaks — Ridgeline", make_terrain, "#00ccff"),
    ("Ripple Function — Ridgeline", make_wave, "#ff6600"),
]

for ax, (title, fn, color) in zip(axes, datasets):
    X, Y, Z = fn()
    n_rows = Z.shape[0]
    ax.set_facecolor("#0a0a0a")

    vertical_spacing = 0.35
    scale = 2.5

    # Draw from back to front for proper occlusion
    for i in range(n_rows):
        row = Z[i, :]
        baseline = i * vertical_spacing
        y_vals = baseline + row * scale

        # Fill below the line with background color for occlusion
        ax.fill_between(
            X[0, :], baseline, y_vals, color="#0a0a0a", zorder=i * 2
        )
        # Draw the line
        alpha = 0.3 + 0.7 * (i / n_rows)
        ax.plot(
            X[0, :],
            y_vals,
            color=color,
            linewidth=0.6,
            alpha=alpha,
            zorder=i * 2 + 1,
        )

    ax.set_xlim(X[0, 0], X[0, -1])
    ax.set_ylim(-1, n_rows * vertical_spacing + Z.max() * scale + 1)
    ax.axis("off")
    ax.set_title(title, color="white", fontsize=14, fontweight="bold", pad=12)

fig.suptitle(
    "Joy Division / Ridgeline — Stacked Cross-Sections",
    color="white",
    fontsize=18,
    fontweight="bold",
    y=0.98,
)
plt.tight_layout(rect=[0, 0, 1, 0.93])
plt.savefig(
    "/home/user/research/isoline-3d-visualization/02_joy_division.png",
    dpi=150,
    bbox_inches="tight",
    facecolor=fig.get_facecolor(),
)
plt.close()
print("Saved 02_joy_division.png")
