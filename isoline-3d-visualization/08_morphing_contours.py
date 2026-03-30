"""
Approach 8: Animated Morphing Contours

Smoothly interpolates between different 3D surfaces, showing how
their contour maps transform. The animation reveals how topology
changes as the underlying surface morphs.
"""

import numpy as np
import matplotlib.pyplot as plt
from PIL import Image
import io


def surface_a(X, Y):
    """Single central peak."""
    return 2.5 * np.exp(-(X**2 + Y**2) / 1.5)


def surface_b(X, Y):
    """Four peaks in a square."""
    Z = np.zeros_like(X)
    for sx, sy in [(1, 1), (1, -1), (-1, 1), (-1, -1)]:
        Z += 1.5 * np.exp(-((X - sx * 1.2) ** 2 + (Y - sy * 1.2) ** 2) / 0.5)
    return Z


def surface_c(X, Y):
    """Ring peak."""
    R = np.sqrt(X**2 + Y**2)
    return 2.0 * np.exp(-((R - 1.5) ** 2) / 0.3)


def surface_d(X, Y):
    """Saddle."""
    return X**2 - Y**2


n = 400
x = np.linspace(-3, 3, n)
y = np.linspace(-3, 3, n)
X, Y = np.meshgrid(x, y)

# Compute all surfaces
Za = surface_a(X, Y)
Zb = surface_b(X, Y)
Zc = surface_c(X, Y)
Zd = surface_d(X, Y)

# Morph sequence: A -> B -> C -> D -> A
surfaces_seq = [Za, Zb, Zc, Zd, Za]
labels = [
    "Single Peak", "Four Peaks", "Ring Peak", "Saddle", "Single Peak"
]
n_transitions = len(surfaces_seq) - 1
frames_per_transition = 15
n_frames = n_transitions * frames_per_transition

frames = []
colors = ["#00ccff", "#ff6644", "#44ff88", "#ffaa00"]

for frame_i in range(n_frames):
    transition = frame_i // frames_per_transition
    t = (frame_i % frames_per_transition) / frames_per_transition

    # Smooth easing
    t_smooth = t * t * (3 - 2 * t)

    Z_from = surfaces_seq[transition]
    Z_to = surfaces_seq[transition + 1]
    Z = Z_from * (1 - t_smooth) + Z_to * t_smooth

    color_from = colors[transition % len(colors)]
    color_to = colors[(transition + 1) % len(colors)]

    fig, ax = plt.subplots(figsize=(8, 8))
    fig.patch.set_facecolor("#0a0a0a")
    ax.set_facecolor("#0a0a0a")
    ax.set_aspect("equal")

    n_levels = 20
    levels = np.linspace(-4, 4, n_levels)

    # Subtle fill
    ax.contourf(X, Y, Z, levels=levels, cmap="magma", alpha=0.1)

    # Isolines with varying thickness for depth
    for i, level in enumerate(levels):
        frac = i / (len(levels) - 1)
        lw = 0.3 + 1.5 * frac
        alpha = 0.2 + 0.7 * frac
        ax.contour(
            X, Y, Z, levels=[level],
            colors=[color_from], linewidths=lw, alpha=alpha,
        )

    label_from = labels[transition]
    label_to = labels[min(transition + 1, len(labels) - 1)]
    ax.set_title(
        f"{label_from} → {label_to}  ({int(t * 100)}%)",
        color="white", fontsize=14, fontweight="bold", pad=12,
    )
    ax.axis("off")

    buf = io.BytesIO()
    plt.savefig(buf, format="png", dpi=80, facecolor=fig.get_facecolor(), bbox_inches="tight")
    buf.seek(0)
    frames.append(Image.open(buf).copy())
    plt.close()

    if frame_i % 10 == 0:
        print(f"  Frame {frame_i + 1}/{n_frames}")

# Save as GIF
frames[0].save(
    "/home/user/research/isoline-3d-visualization/08_morphing_contours.gif",
    save_all=True,
    append_images=frames[1:],
    duration=100,
    loop=0,
)
print("Saved 08_morphing_contours.gif")
