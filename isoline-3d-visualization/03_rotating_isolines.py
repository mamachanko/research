"""
Approach 3: Rotating Isoline Animation

Projects a 3D surface into 2D from multiple viewing angles,
drawing contour lines at each frame to produce an animated GIF
that reveals the 3D shape through rotation.
"""

import numpy as np
import matplotlib.pyplot as plt
from mpl_toolkits.mplot3d import Axes3D
from PIL import Image
import io


def make_surface():
    u = np.linspace(0, 2 * np.pi, 200)
    v = np.linspace(0, np.pi, 100)
    U, V = np.meshgrid(u, v)

    # Bumpy sphere
    r = 1 + 0.3 * np.sin(4 * U) * np.sin(3 * V)
    X = r * np.sin(V) * np.cos(U)
    Y = r * np.sin(V) * np.sin(U)
    Z = r * np.cos(V)
    return X, Y, Z


def make_torus():
    u = np.linspace(0, 2 * np.pi, 200)
    v = np.linspace(0, 2 * np.pi, 100)
    U, V = np.meshgrid(u, v)
    R, r = 2.0, 0.8
    X = (R + r * np.cos(V)) * np.cos(U)
    Y = (R + r * np.cos(V)) * np.sin(U)
    Z = r * np.sin(V)
    return X, Y, Z


n_frames = 36
frames = []

for frame_i in range(n_frames):
    angle = frame_i * 360 / n_frames

    fig = plt.figure(figsize=(14, 6))
    fig.patch.set_facecolor("#0a0a0a")

    surfaces = [
        ("Bumpy Sphere", make_surface, "#00ffaa"),
        ("Torus", make_torus, "#ff44aa"),
    ]

    for idx, (title, fn, color) in enumerate(surfaces):
        ax = fig.add_subplot(1, 2, idx + 1, projection="3d")
        ax.set_facecolor("#0a0a0a")

        X, Y, Z = fn()

        # Draw isolines in three directions
        n_lines = 16

        # Constant-u lines
        step = X.shape[1] // n_lines
        for j in range(0, X.shape[1], step):
            ax.plot(
                X[:, j], Y[:, j], Z[:, j],
                color=color, linewidth=0.5, alpha=0.7,
            )

        # Constant-v lines
        step = X.shape[0] // n_lines
        for i in range(0, X.shape[0], step):
            ax.plot(
                X[i, :], Y[i, :], Z[i, :],
                color=color, linewidth=0.5, alpha=0.7,
            )

        ax.view_init(elev=25, azim=angle)
        ax.set_xlim(-3, 3)
        ax.set_ylim(-3, 3)
        ax.set_zlim(-2, 2)
        ax.axis("off")
        ax.set_title(title, color="white", fontsize=13, fontweight="bold", pad=0)

    fig.suptitle(
        "Rotating Isoline View",
        color="white", fontsize=16, fontweight="bold", y=0.95,
    )
    plt.tight_layout(rect=[0, 0, 1, 0.92])

    buf = io.BytesIO()
    plt.savefig(buf, format="png", dpi=100, facecolor=fig.get_facecolor(), bbox_inches="tight")
    buf.seek(0)
    frames.append(Image.open(buf).copy())
    plt.close()

    if frame_i % 6 == 0:
        print(f"  Frame {frame_i + 1}/{n_frames}")

# Save as GIF
frames[0].save(
    "/home/user/research/isoline-3d-visualization/03_rotating_isolines.gif",
    save_all=True,
    append_images=frames[1:],
    duration=80,
    loop=0,
)
print("Saved 03_rotating_isolines.gif")
