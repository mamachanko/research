"""
Approach 7: Parametric Surface Isolines

Draws iso-parameter curves (constant-u and constant-v lines) on
parametric surfaces, revealing shape through the natural coordinate
grid of the surface itself.
"""

import numpy as np
import matplotlib.pyplot as plt
from mpl_toolkits.mplot3d import Axes3D


def klein_bottle_figure8(u, v):
    """Figure-8 Klein bottle immersion in 3D."""
    a = 2
    x = (a + np.cos(u / 2) * np.sin(v) - np.sin(u / 2) * np.sin(2 * v)) * np.cos(u)
    y = (a + np.cos(u / 2) * np.sin(v) - np.sin(u / 2) * np.sin(2 * v)) * np.sin(u)
    z = np.sin(u / 2) * np.sin(v) + np.cos(u / 2) * np.sin(2 * v)
    return x, y, z


def mobius_strip(u, v):
    """Mobius strip parameterization."""
    x = (1 + v / 2 * np.cos(u / 2)) * np.cos(u)
    y = (1 + v / 2 * np.cos(u / 2)) * np.sin(u)
    z = v / 2 * np.sin(u / 2)
    return x, y, z


def trefoil_tube(u, v):
    """Tube around a trefoil knot."""
    r = 0.3
    # Trefoil knot
    cx = np.sin(u) + 2 * np.sin(2 * u)
    cy = np.cos(u) - 2 * np.cos(2 * u)
    cz = -np.sin(3 * u)

    # Tangent
    dcx = np.cos(u) + 4 * np.cos(2 * u)
    dcy = -np.sin(u) + 4 * np.sin(2 * u)
    dcz = -3 * np.cos(3 * u)
    norm = np.sqrt(dcx**2 + dcy**2 + dcz**2)
    tx, ty, tz = dcx / norm, dcy / norm, dcz / norm

    # Normal and binormal (approximate)
    nx = -ty
    ny = tx
    nz = np.zeros_like(tx)
    n_norm = np.sqrt(nx**2 + ny**2 + nz**2) + 1e-8
    nx, ny, nz = nx / n_norm, ny / n_norm, nz / n_norm

    bx = ty * nz - tz * ny
    by = tz * nx - tx * nz
    bz = tx * ny - ty * nx

    x = cx + r * (np.cos(v) * nx + np.sin(v) * bx)
    y = cy + r * (np.cos(v) * ny + np.sin(v) * by)
    z = cz + r * (np.cos(v) * nz + np.sin(v) * bz)
    return x, y, z


surfaces = [
    ("Klein Bottle (Figure-8)", klein_bottle_figure8,
     (0, 2 * np.pi), (0, 2 * np.pi), 30, 30, "#ff9944", "#44ccff"),
    ("Mobius Strip", mobius_strip,
     (0, 2 * np.pi), (-1, 1), 40, 8, "#44ff99", "#ff44cc"),
    ("Trefoil Knot Tube", trefoil_tube,
     (0, 2 * np.pi), (0, 2 * np.pi), 60, 12, "#ffdd44", "#ff4466"),
]

fig = plt.figure(figsize=(18, 6))
fig.patch.set_facecolor("#0a0a0a")

for idx, (title, fn, u_range, v_range, nu, nv, col_u, col_v) in enumerate(surfaces):
    ax = fig.add_subplot(1, 3, idx + 1, projection="3d")
    ax.set_facecolor("#0a0a0a")

    u_full = np.linspace(*u_range, 300)
    v_full = np.linspace(*v_range, 300)

    # Constant-u lines
    u_vals = np.linspace(*u_range, nu, endpoint=False)
    for u_val in u_vals:
        u_arr = np.full_like(v_full, u_val)
        x, y, z = fn(u_arr, v_full)
        ax.plot(x, y, z, color=col_u, linewidth=0.4, alpha=0.6)

    # Constant-v lines
    v_vals = np.linspace(*v_range, nv, endpoint=False)
    for v_val in v_vals:
        v_arr = np.full_like(u_full, v_val)
        x, y, z = fn(u_full, v_arr)
        ax.plot(x, y, z, color=col_v, linewidth=0.4, alpha=0.6)

    ax.view_init(elev=25, azim=45)
    ax.axis("off")
    ax.set_title(title, color="white", fontsize=13, fontweight="bold", pad=5)

fig.suptitle(
    "Parametric Surface Isolines — Constant-u & Constant-v Curves",
    color="white", fontsize=17, fontweight="bold", y=0.98,
)
plt.tight_layout(rect=[0, 0, 1, 0.92])
plt.savefig(
    "/home/user/research/isoline-3d-visualization/07_parametric_surfaces.png",
    dpi=150, bbox_inches="tight", facecolor=fig.get_facecolor(),
)
plt.close()
print("Saved 07_parametric_surfaces.png")
