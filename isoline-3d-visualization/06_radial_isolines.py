"""
Approach 6: Radial Isolines

Contour lines based on distance from center, combined with angular
and height information. Particularly effective for revealing spherical,
toroidal, and axially-symmetric objects.
"""

import numpy as np
import matplotlib.pyplot as plt


def sphere_projection(n=500):
    """Project a sphere's radial contours onto 2D."""
    theta = np.linspace(0, 2 * np.pi, n)
    n_rings = 20

    fig_data = []
    for i in range(1, n_rings + 1):
        phi = i * np.pi / (n_rings + 1)
        r = np.sin(phi)  # radius of ring at this latitude
        z = np.cos(phi)  # height
        x = r * np.cos(theta)
        y = r * np.sin(theta)
        fig_data.append((x, y, z))
    return fig_data


def torus_projection(n=500):
    """Project torus iso-curves onto 2D."""
    theta = np.linspace(0, 2 * np.pi, n)
    R, r = 2.0, 0.8
    n_rings = 24

    fig_data = []
    # Toroidal circles (around the tube)
    for i in range(n_rings):
        u = i * 2 * np.pi / n_rings
        cx = R * np.cos(u)
        cy = R * np.sin(u)
        x = cx + r * np.cos(theta) * np.cos(u)
        y = cy + r * np.cos(theta) * np.sin(u)
        z = r * np.sin(theta)
        fig_data.append((x, y, z, "toroidal"))

    # Poloidal circles (along the ring)
    for i in range(12):
        v = i * 2 * np.pi / 12
        x = (R + r * np.cos(v)) * np.cos(theta)
        y = (R + r * np.cos(v)) * np.sin(theta)
        z_val = r * np.sin(v)
        fig_data.append((x, y, np.full_like(theta, z_val), "poloidal"))

    return fig_data


fig, axes = plt.subplots(1, 2, figsize=(16, 8))
fig.patch.set_facecolor("#0a0a0a")

# Sphere
ax = axes[0]
ax.set_facecolor("#0a0a0a")
ax.set_aspect("equal")
data = sphere_projection()
for x, y, z in data:
    t = (z + 1) / 2  # normalize z to 0-1
    alpha = 0.3 + 0.6 * (1 - abs(z))
    lw = 0.4 + 1.5 * (1 - abs(z))
    color = plt.cm.cool(t)
    ax.plot(x, y, color=color, linewidth=lw, alpha=alpha)

# Add meridian lines
theta = np.linspace(0, np.pi, 200)
for i in range(12):
    phi = i * np.pi / 6
    x = np.sin(theta) * np.cos(phi)
    y = np.sin(theta) * np.sin(phi)
    z = np.cos(theta)
    colors = plt.cm.cool((z + 1) / 2)
    for j in range(len(theta) - 1):
        ax.plot(
            [x[j], x[j + 1]], [y[j], y[j + 1]],
            color=colors[j], linewidth=0.4, alpha=0.5,
        )

ax.set_xlim(-1.3, 1.3)
ax.set_ylim(-1.3, 1.3)
ax.axis("off")
ax.set_title("Sphere — Latitude + Meridian Isolines", color="white",
             fontsize=13, fontweight="bold", pad=12)

# Torus
ax = axes[1]
ax.set_facecolor("#0a0a0a")
ax.set_aspect("equal")
data = torus_projection()
for item in data:
    if item[3] == "toroidal":
        x, y, z, _ = item
        color = "#ff6688"
        ax.plot(x, y, color=color, linewidth=0.4, alpha=0.5)
    else:
        x, y, z, _ = item
        color = "#66ddff"
        ax.plot(x, y, color=color, linewidth=0.7, alpha=0.7)

ax.set_xlim(-3.5, 3.5)
ax.set_ylim(-3.5, 3.5)
ax.axis("off")
ax.set_title("Torus — Toroidal + Poloidal Isolines", color="white",
             fontsize=13, fontweight="bold", pad=12)

fig.suptitle(
    "Radial Isolines — Distance & Angular Contours",
    color="white", fontsize=18, fontweight="bold", y=0.98,
)
plt.tight_layout(rect=[0, 0, 1, 0.93])
plt.savefig(
    "/home/user/research/isoline-3d-visualization/06_radial_isolines.png",
    dpi=150, bbox_inches="tight", facecolor=fig.get_facecolor(),
)
plt.close()
print("Saved 06_radial_isolines.png")
