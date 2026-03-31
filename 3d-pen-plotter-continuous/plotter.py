#!/usr/bin/env python3
"""
3D Pen Plotter Continuous Stroke — Single continuous line drawings of 3D objects.

Pipeline:
1. Define 3D geometry as vertices + edges + faces
2. Apply rotation and perspective projection
3. Eulerize the edge graph (Chinese Postman Problem) so a continuous path exists
4. Find Euler circuit via Hierholzer's algorithm
5. Render to SVG with depth-varying line weight for 3D perception

Dependencies: Python 3.8+ standard library only (no external packages).
"""

import math
import itertools
import xml.etree.ElementTree as ET
from dataclasses import dataclass, field
from typing import List, Tuple, Dict, Set, Optional
from collections import defaultdict
from copy import deepcopy

# ─── Vector / Matrix Helpers ────────────────────────────────────────────────

Vec3 = Tuple[float, float, float]
Vec2 = Tuple[float, float]


def v3_add(a: Vec3, b: Vec3) -> Vec3:
    return (a[0]+b[0], a[1]+b[1], a[2]+b[2])

def v3_sub(a: Vec3, b: Vec3) -> Vec3:
    return (a[0]-b[0], a[1]-b[1], a[2]-b[2])

def v3_scale(a: Vec3, s: float) -> Vec3:
    return (a[0]*s, a[1]*s, a[2]*s)

def v3_dot(a: Vec3, b: Vec3) -> float:
    return a[0]*b[0] + a[1]*b[1] + a[2]*b[2]

def v3_cross(a: Vec3, b: Vec3) -> Vec3:
    return (a[1]*b[2]-a[2]*b[1], a[2]*b[0]-a[0]*b[2], a[0]*b[1]-a[1]*b[0])

def v3_length(a: Vec3) -> float:
    return math.sqrt(v3_dot(a, a))

def v3_normalize(a: Vec3) -> Vec3:
    l = v3_length(a)
    if l < 1e-12:
        return (0.0, 0.0, 0.0)
    return v3_scale(a, 1.0/l)


def rotate_x(v: Vec3, angle: float) -> Vec3:
    c, s = math.cos(angle), math.sin(angle)
    return (v[0], c*v[1] - s*v[2], s*v[1] + c*v[2])

def rotate_y(v: Vec3, angle: float) -> Vec3:
    c, s = math.cos(angle), math.sin(angle)
    return (c*v[0] + s*v[2], v[1], -s*v[0] + c*v[2])

def rotate_z(v: Vec3, angle: float) -> Vec3:
    c, s = math.cos(angle), math.sin(angle)
    return (c*v[0] - s*v[1], s*v[0] + c*v[1], v[2])


# ─── 3D Geometry Definitions ────────────────────────────────────────────────

@dataclass
class Mesh:
    """A 3D wireframe mesh: vertices, edges, and optional faces."""
    name: str
    vertices: List[Vec3]
    edges: List[Tuple[int, int]]
    faces: List[List[int]] = field(default_factory=list)  # vertex indices per face


def make_cube(size: float = 1.0) -> Mesh:
    s = size / 2
    verts = [
        (-s, -s, -s), ( s, -s, -s), ( s,  s, -s), (-s,  s, -s),
        (-s, -s,  s), ( s, -s,  s), ( s,  s,  s), (-s,  s,  s),
    ]
    edges = [
        (0,1),(1,2),(2,3),(3,0),  # back face
        (4,5),(5,6),(6,7),(7,4),  # front face
        (0,4),(1,5),(2,6),(3,7),  # connecting
    ]
    faces = [
        [0,1,2,3], [4,5,6,7], [0,1,5,4],
        [2,3,7,6], [0,3,7,4], [1,2,6,5],
    ]
    return Mesh("cube", verts, edges, faces)


def make_tetrahedron(size: float = 1.0) -> Mesh:
    s = size / math.sqrt(2)
    verts = [(s,s,s), (s,-s,-s), (-s,s,-s), (-s,-s,s)]
    edges = [(0,1),(0,2),(0,3),(1,2),(1,3),(2,3)]
    faces = [[0,1,2],[0,1,3],[0,2,3],[1,2,3]]
    return Mesh("tetrahedron", verts, edges, faces)


def make_octahedron(size: float = 1.0) -> Mesh:
    s = size
    verts = [(s,0,0),(-s,0,0),(0,s,0),(0,-s,0),(0,0,s),(0,0,-s)]
    edges = [
        (0,2),(0,3),(0,4),(0,5),
        (1,2),(1,3),(1,4),(1,5),
        (2,4),(2,5),(3,4),(3,5),
    ]
    faces = [
        [0,2,4],[0,4,3],[0,3,5],[0,5,2],
        [1,2,4],[1,4,3],[1,3,5],[1,5,2],
    ]
    return Mesh("octahedron", verts, edges, faces)


def make_icosahedron(size: float = 1.0) -> Mesh:
    phi = (1 + math.sqrt(5)) / 2
    s = size / math.sqrt(1 + phi*phi)
    verts = [
        (-s,  phi*s, 0), ( s,  phi*s, 0), (-s, -phi*s, 0), ( s, -phi*s, 0),
        (0, -s,  phi*s), (0,  s,  phi*s), (0, -s, -phi*s), (0,  s, -phi*s),
        ( phi*s, 0, -s), ( phi*s, 0,  s), (-phi*s, 0, -s), (-phi*s, 0,  s),
    ]
    face_indices = [
        [0,11,5],[0,5,1],[0,1,7],[0,7,10],[0,10,11],
        [1,5,9],[5,11,4],[11,10,2],[10,7,6],[7,1,8],
        [3,9,4],[3,4,2],[3,2,6],[3,6,8],[3,8,9],
        [4,9,5],[2,4,11],[6,2,10],[8,6,7],[9,8,1],
    ]
    edge_set = set()
    for f in face_indices:
        for i in range(3):
            a, b = f[i], f[(i+1)%3]
            edge_set.add((min(a,b), max(a,b)))
    return Mesh("icosahedron", verts, list(edge_set), face_indices)


def make_dodecahedron(size: float = 1.0) -> Mesh:
    phi = (1 + math.sqrt(5)) / 2
    iphi = 1.0 / phi
    s = size * 0.6
    verts = [
        # cube vertices
        ( s, s, s),( s, s,-s),( s,-s, s),( s,-s,-s),
        (-s, s, s),(-s, s,-s),(-s,-s, s),(-s,-s,-s),
        # rectangle vertices
        (0,  s*phi,  s*iphi), (0,  s*phi, -s*iphi),
        (0, -s*phi,  s*iphi), (0, -s*phi, -s*iphi),
        ( s*iphi, 0,  s*phi), (-s*iphi, 0,  s*phi),
        ( s*iphi, 0, -s*phi), (-s*iphi, 0, -s*phi),
        ( s*phi,  s*iphi, 0), ( s*phi, -s*iphi, 0),
        (-s*phi,  s*iphi, 0), (-s*phi, -s*iphi, 0),
    ]
    face_indices = [
        [0,16,17,2,12],[0,12,13,4,8],[0,8,9,1,16],
        [1,9,5,15,14],[1,14,3,17,16],[2,17,3,11,10],
        [2,10,6,13,12],[4,13,6,19,18],[4,18,5,9,8],
        [7,11,3,14,15],[7,15,5,18,19],[7,19,6,10,11],
    ]
    edge_set = set()
    for f in face_indices:
        n = len(f)
        for i in range(n):
            a, b = f[i], f[(i+1)%n]
            edge_set.add((min(a,b), max(a,b)))
    return Mesh("dodecahedron", verts, list(edge_set), face_indices)


def make_torus(R: float = 1.0, r: float = 0.4, n_major: int = 12, n_minor: int = 8) -> Mesh:
    """Torus wireframe: R=major radius, r=minor radius."""
    verts = []
    for i in range(n_major):
        theta = 2 * math.pi * i / n_major
        for j in range(n_minor):
            phi = 2 * math.pi * j / n_minor
            x = (R + r * math.cos(phi)) * math.cos(theta)
            y = (R + r * math.cos(phi)) * math.sin(theta)
            z = r * math.sin(phi)
            verts.append((x, y, z))
    edges = []
    for i in range(n_major):
        for j in range(n_minor):
            idx = i * n_minor + j
            # ring edge
            next_j = i * n_minor + (j + 1) % n_minor
            edges.append((idx, next_j))
            # tube edge
            next_i = ((i + 1) % n_major) * n_minor + j
            edges.append((idx, next_i))
    return Mesh("torus", verts, edges, [])


def make_sphere_wireframe(radius: float = 1.0, n_lat: int = 8, n_lon: int = 12) -> Mesh:
    """Sphere wireframe from latitude/longitude lines."""
    verts = [(0, 0, radius)]  # north pole
    for i in range(1, n_lat):
        lat = math.pi * i / n_lat
        for j in range(n_lon):
            lon = 2 * math.pi * j / n_lon
            x = radius * math.sin(lat) * math.cos(lon)
            y = radius * math.sin(lat) * math.sin(lon)
            z = radius * math.cos(lat)
            verts.append((x, y, z))
    verts.append((0, 0, -radius))  # south pole
    south = len(verts) - 1

    edges = []
    # connect north pole to first ring
    for j in range(n_lon):
        edges.append((0, 1 + j))
    # latitude rings and longitude connections
    for i in range(1, n_lat):
        ring_start = 1 + (i - 1) * n_lon
        for j in range(n_lon):
            # latitude edge (ring)
            edges.append((ring_start + j, ring_start + (j + 1) % n_lon))
            # longitude edge (to next ring or south pole)
            if i < n_lat - 1:
                next_ring = ring_start + n_lon
                edges.append((ring_start + j, next_ring + j))
            else:
                edges.append((ring_start + j, south))
    return Mesh("sphere", verts, edges, [])


def make_stellated_octahedron(size: float = 1.0) -> Mesh:
    """Stella octangula — two interlocking tetrahedra."""
    s = size
    # First tetrahedron
    t1 = [(s,s,s),(s,-s,-s),(-s,s,-s),(-s,-s,s)]
    # Second tetrahedron (inverted)
    t2 = [(-s,-s,-s),(-s,s,s),(s,-s,s),(s,s,-s)]
    verts = t1 + t2
    edges = [
        (0,1),(0,2),(0,3),(1,2),(1,3),(2,3),  # first tetra
        (4,5),(4,6),(4,7),(5,6),(5,7),(6,7),  # second tetra
    ]
    return Mesh("stella_octangula", verts, edges, [])


def make_prism(n_sides: int = 6, radius: float = 1.0, height: float = 1.5) -> Mesh:
    """Regular n-gon prism."""
    verts = []
    for i in range(n_sides):
        angle = 2 * math.pi * i / n_sides
        x, y = radius * math.cos(angle), radius * math.sin(angle)
        verts.append((x, y, -height/2))
        verts.append((x, y,  height/2))
    edges = []
    for i in range(n_sides):
        # vertical edge
        edges.append((2*i, 2*i+1))
        # bottom ring
        edges.append((2*i, 2*((i+1) % n_sides)))
        # top ring
        edges.append((2*i+1, 2*((i+1) % n_sides)+1))
    return Mesh(f"{n_sides}-prism", verts, edges, [])


def make_moebius_strip(n_segments: int = 30, width: float = 0.4, radius: float = 1.0) -> Mesh:
    """Möbius strip wireframe — a non-orientable surface."""
    verts = []
    n_width = 4  # cross-section resolution
    for i in range(n_segments):
        t = 2 * math.pi * i / n_segments
        half_twist = t / 2
        cx = radius * math.cos(t)
        cy = radius * math.sin(t)
        # tangent direction along the strip
        for j in range(n_width):
            s = width * (j / (n_width - 1) - 0.5)
            x = cx + s * math.cos(half_twist) * math.cos(t)
            y = cy + s * math.cos(half_twist) * math.sin(t)
            z = s * math.sin(half_twist)
            verts.append((x, y, z))
    edges = []
    for i in range(n_segments):
        base = i * n_width
        next_base = ((i + 1) % n_segments) * n_width
        for j in range(n_width):
            # along strip
            if i < n_segments - 1:
                edges.append((base + j, next_base + j))
            else:
                # Möbius twist: connect last to first reversed
                edges.append((base + j, next_base + (n_width - 1 - j)))
            # across strip
            if j < n_width - 1:
                edges.append((base + j, base + j + 1))
    return Mesh("moebius_strip", verts, edges, [])


# ─── 3D to 2D Projection ────────────────────────────────────────────────────

@dataclass
class Camera:
    position: Vec3 = (3.0, 2.0, 4.0)
    target: Vec3 = (0.0, 0.0, 0.0)
    up: Vec3 = (0.0, 0.0, 1.0)  # Z-up
    fov: float = 60.0  # degrees
    width: float = 800.0
    height: float = 800.0


@dataclass
class ProjectedVertex:
    screen: Vec2
    depth: float  # camera-space Z (larger = farther)


def project_mesh(mesh: Mesh, camera: Camera,
                 rot_x: float = 0, rot_y: float = 0, rot_z: float = 0
                 ) -> Tuple[List[ProjectedVertex], float, float]:
    """Project all vertices. Returns projected verts and (z_near, z_far) range."""
    # Apply rotation to vertices
    rotated = []
    for v in mesh.vertices:
        v2 = rotate_x(v, rot_x)
        v2 = rotate_y(v2, rot_y)
        v2 = rotate_z(v2, rot_z)
        rotated.append(v2)

    # Build view transform
    fwd = v3_normalize(v3_sub(camera.target, camera.position))
    right = v3_normalize(v3_cross(fwd, camera.up))
    up = v3_cross(right, fwd)

    focal = 1.0 / math.tan(math.radians(camera.fov / 2))
    cx, cy = camera.width / 2, camera.height / 2
    scale = min(camera.width, camera.height) / 2 * focal

    projected = []
    z_vals = []
    for v in rotated:
        # Transform to camera space
        rel = v3_sub(v, camera.position)
        cam_x = v3_dot(rel, right)
        cam_y = v3_dot(rel, up)
        cam_z = v3_dot(rel, fwd)  # depth along view direction

        if cam_z < 0.01:
            cam_z = 0.01  # clamp near plane

        # Perspective divide
        sx = scale * (cam_x / cam_z) + cx
        sy = cy - scale * (cam_y / cam_z)  # flip Y for screen coords
        projected.append(ProjectedVertex((sx, sy), cam_z))
        z_vals.append(cam_z)

    z_near = min(z_vals)
    z_far = max(z_vals)
    return projected, z_near, z_far


# ─── Graph Eulerization (Chinese Postman) ────────────────────────────────────

class Graph:
    """Multigraph for Eulerization and Euler circuit finding."""

    def __init__(self, n_vertices: int):
        self.n = n_vertices
        # adjacency: vertex -> list of (neighbor, edge_id)
        self.adj: Dict[int, List[Tuple[int, int]]] = defaultdict(list)
        self.edges: List[Tuple[int, int]] = []
        self.edge_used: List[bool] = []

    def add_edge(self, u: int, v: int) -> int:
        eid = len(self.edges)
        self.edges.append((u, v))
        self.edge_used.append(False)
        self.adj[u].append((v, eid))
        self.adj[v].append((u, eid))
        return eid

    def degree(self, v: int) -> int:
        return len(self.adj[v])

    def odd_degree_vertices(self) -> List[int]:
        return [v for v in range(self.n) if self.degree(v) % 2 == 1]

    def bfs_shortest_path(self, start: int, end: int) -> List[int]:
        """BFS shortest path returning list of vertices."""
        from collections import deque
        visited = {start: None}
        queue = deque([start])
        while queue:
            v = queue.popleft()
            if v == end:
                path = []
                while v is not None:
                    path.append(v)
                    v = visited[v]
                return path[::-1]
            for (u, _) in self.adj[v]:
                if u not in visited:
                    visited[u] = v
                    queue.append(u)
        return []  # disconnected

    def eulerize(self) -> 'Graph':
        """
        Chinese Postman: add minimum duplicate edges to make all vertices even-degree.
        Returns a new Graph with the additional edges.
        """
        odd = self.odd_degree_vertices()
        if not odd:
            return self  # already Eulerian

        # Compute shortest paths between all pairs of odd-degree vertices
        n_odd = len(odd)
        dist = [[0]*n_odd for _ in range(n_odd)]
        paths = [[[] for _ in range(n_odd)] for _ in range(n_odd)]

        for i in range(n_odd):
            for j in range(i+1, n_odd):
                p = self.bfs_shortest_path(odd[i], odd[j])
                d = len(p) - 1 if p else float('inf')
                dist[i][j] = d
                dist[j][i] = d
                paths[i][j] = p
                paths[j][i] = p[::-1]

        # Minimum weight perfect matching (brute force — fine for small odd sets)
        matching = self._min_perfect_matching(list(range(n_odd)), dist)

        # Build new graph with duplicated edges
        new_graph = Graph(self.n)
        for u, v in self.edges:
            new_graph.add_edge(u, v)

        for i, j in matching:
            path = paths[i][j]
            for k in range(len(path) - 1):
                new_graph.add_edge(path[k], path[k+1])

        return new_graph

    @staticmethod
    def _min_perfect_matching(vertices: List[int], dist) -> List[Tuple[int, int]]:
        """Minimum weight perfect matching. Uses bitmask DP for small n, greedy for large n."""
        n = len(vertices)
        if n <= 1:
            return []
        if n == 2:
            return [(vertices[0], vertices[1])]

        # For large vertex sets, use greedy nearest-neighbor matching
        if n > 22:
            return Graph._greedy_matching(vertices, dist)

        full = (1 << n) - 1
        INF = float('inf')
        dp = [INF] * (1 << n)
        parent = [(-1, -1)] * (1 << n)
        dp[0] = 0

        for mask in range(1 << n):
            if dp[mask] == INF:
                continue
            first = -1
            for b in range(n):
                if not (mask & (1 << b)):
                    first = b
                    break
            if first == -1:
                continue
            for second in range(first + 1, n):
                if mask & (1 << second):
                    continue
                new_mask = mask | (1 << first) | (1 << second)
                cost = dp[mask] + dist[vertices[first]][vertices[second]]
                if cost < dp[new_mask]:
                    dp[new_mask] = cost
                    parent[new_mask] = (first, second)

        matching = []
        mask = full
        while mask:
            i, j = parent[mask]
            matching.append((vertices[i], vertices[j]))
            mask ^= (1 << i) | (1 << j)
        return matching

    @staticmethod
    def _greedy_matching(vertices: List[int], dist) -> List[Tuple[int, int]]:
        """Greedy nearest-neighbor matching for large odd-vertex sets."""
        remaining = set(vertices)
        matching = []
        while len(remaining) >= 2:
            best_d = float('inf')
            best_pair = None
            # Pick the vertex with the closest unmatched neighbor
            for u in remaining:
                for v in remaining:
                    if u >= v:
                        continue
                    d = dist[u][v]
                    if d < best_d:
                        best_d = d
                        best_pair = (u, v)
            if best_pair:
                matching.append(best_pair)
                remaining.discard(best_pair[0])
                remaining.discard(best_pair[1])
        return matching

    def euler_circuit(self, start: int = 0) -> List[int]:
        """Hierholzer's algorithm. Returns vertex sequence."""
        # Reset edge usage
        for i in range(len(self.edge_used)):
            self.edge_used[i] = False

        stack = [start]
        circuit = []

        while stack:
            v = stack[-1]
            found = False
            for (u, eid) in self.adj[v]:
                if not self.edge_used[eid]:
                    self.edge_used[eid] = True
                    stack.append(u)
                    found = True
                    break
            if not found:
                stack.pop()
                circuit.append(v)

        return circuit[::-1]


def _connected_components(n_verts: int, edges: List[Tuple[int, int]]) -> List[Set[int]]:
    """Find connected components via union-find."""
    parent = list(range(n_verts))

    def find(x):
        while parent[x] != x:
            parent[x] = parent[parent[x]]
            x = parent[x]
        return x

    def union(a, b):
        a, b = find(a), find(b)
        if a != b:
            parent[a] = b

    for u, v in edges:
        union(u, v)

    components = defaultdict(set)
    for i in range(n_verts):
        # Only include vertices that have at least one edge
        components[find(i)].add(i)

    # Filter to components that have edges
    edge_verts = set()
    for u, v in edges:
        edge_verts.add(u)
        edge_verts.add(v)

    result = []
    seen_roots = set()
    for v in edge_verts:
        r = find(v)
        if r not in seen_roots:
            seen_roots.add(r)
            result.append({u for u in components[r] if u in edge_verts})
    return result


def _connect_components(mesh: Mesh) -> List[Tuple[int, int]]:
    """Add bridge edges between disconnected components (closest vertex pairs)."""
    components = _connected_components(len(mesh.vertices), mesh.edges)
    if len(components) <= 1:
        return []

    extra_edges = []
    # Greedily connect components by nearest vertex pairs
    connected = [components[0]]
    remaining = list(components[1:])

    while remaining:
        best_dist = float('inf')
        best_pair = None
        best_idx = 0

        for ci, comp in enumerate(remaining):
            for v in comp:
                for cc in connected:
                    for u in cc:
                        d = v3_length(v3_sub(mesh.vertices[u], mesh.vertices[v]))
                        if d < best_dist:
                            best_dist = d
                            best_pair = (u, v)
                            best_idx = ci

        if best_pair:
            # Add two edges (there and back) to maintain parity awareness
            extra_edges.append(best_pair)
            connected[0] = connected[0] | remaining[best_idx]
            remaining.pop(best_idx)

    return extra_edges


def mesh_to_euler_path(mesh: Mesh) -> List[int]:
    """Convert mesh edges into a single continuous vertex path."""
    # First, connect any disconnected components
    bridge_edges = _connect_components(mesh)

    all_edges = list(mesh.edges) + bridge_edges

    g = Graph(len(mesh.vertices))
    for u, v in all_edges:
        g.add_edge(u, v)
    g2 = g.eulerize()
    circuit = g2.euler_circuit(start=0)
    return circuit


# ─── SVG Rendering with Depth Cues ──────────────────────────────────────────

@dataclass
class RenderConfig:
    width: float = 800
    height: float = 800
    padding: float = 60
    stroke_min: float = 0.4
    stroke_max: float = 3.5
    opacity_min: float = 0.3
    opacity_max: float = 1.0
    stroke_color: str = "#1a1a2e"
    background: str = "#faf9f6"
    show_vertices: bool = False


def depth_to_weight(z: float, z_near: float, z_far: float, w_min: float, w_max: float) -> float:
    if z_far - z_near < 1e-6:
        return (w_min + w_max) / 2
    t = (z - z_near) / (z_far - z_near)  # 0=near, 1=far
    return w_max - t * (w_max - w_min)  # near=thick, far=thin


def depth_to_opacity(z: float, z_near: float, z_far: float, o_min: float, o_max: float) -> float:
    if z_far - z_near < 1e-6:
        return (o_min + o_max) / 2
    t = (z - z_near) / (z_far - z_near)
    return o_max - t * (o_max - o_min)


def render_continuous_svg(
    mesh: Mesh,
    camera: Camera,
    config: RenderConfig,
    rot_x: float = 0, rot_y: float = 0, rot_z: float = 0,
    title: str = "",
) -> str:
    """Render mesh as a single continuous SVG path with depth-varying line weight."""

    projected, z_near, z_far = project_mesh(mesh, camera, rot_x, rot_y, rot_z)

    # Get the Euler circuit (single continuous path)
    circuit = mesh_to_euler_path(mesh)

    if not circuit:
        return "<svg></svg>"

    # Build SVG
    svg = ET.Element("svg", {
        "xmlns": "http://www.w3.org/2000/svg",
        "width": str(int(config.width)),
        "height": str(int(config.height)),
        "viewBox": f"0 0 {int(config.width)} {int(config.height)}",
    })

    # Background
    ET.SubElement(svg, "rect", {
        "width": "100%", "height": "100%",
        "fill": config.background,
    })

    # Title
    if title:
        t = ET.SubElement(svg, "text", {
            "x": str(config.width / 2), "y": "30",
            "text-anchor": "middle",
            "font-family": "monospace",
            "font-size": "16",
            "fill": "#333",
        })
        t.text = title

    # Render path as individual segments with varying stroke width
    # This gives us per-segment depth cues while maintaining continuous drawing
    group = ET.SubElement(svg, "g", {"id": "continuous-path"})

    for i in range(len(circuit) - 1):
        v0_idx = circuit[i]
        v1_idx = circuit[i + 1]
        p0 = projected[v0_idx]
        p1 = projected[v1_idx]

        avg_depth = (p0.depth + p1.depth) / 2
        w = depth_to_weight(avg_depth, z_near, z_far, config.stroke_min, config.stroke_max)
        o = depth_to_opacity(avg_depth, z_near, z_far, config.opacity_min, config.opacity_max)

        ET.SubElement(group, "line", {
            "x1": f"{p0.screen[0]:.2f}",
            "y1": f"{p0.screen[1]:.2f}",
            "x2": f"{p1.screen[0]:.2f}",
            "y2": f"{p1.screen[1]:.2f}",
            "stroke": config.stroke_color,
            "stroke-width": f"{w:.2f}",
            "stroke-opacity": f"{o:.2f}",
            "stroke-linecap": "round",
            "stroke-linejoin": "round",
        })

    # Optional: mark start point
    start_proj = projected[circuit[0]]
    ET.SubElement(group, "circle", {
        "cx": f"{start_proj.screen[0]:.2f}",
        "cy": f"{start_proj.screen[1]:.2f}",
        "r": "4",
        "fill": "#e63946",
        "opacity": "0.8",
    })

    # Stats annotation
    n_orig = len(mesh.edges)
    n_total = len(circuit) - 1
    n_dup = n_total - n_orig
    stats = ET.SubElement(svg, "text", {
        "x": str(config.width / 2),
        "y": str(config.height - 15),
        "text-anchor": "middle",
        "font-family": "monospace",
        "font-size": "11",
        "fill": "#999",
    })
    stats.text = f"{n_orig} edges | {n_dup} retraced | {n_total} segments total | 1 continuous stroke"

    return ET.tostring(svg, encoding="unicode", xml_declaration=False)


# ─── Gallery Generation ──────────────────────────────────────────────────────

def generate_all():
    """Generate SVGs for various 3D objects."""

    shapes = [
        (make_cube(1.5), 0.35, 0.55, 0.2, "Cube"),
        (make_tetrahedron(1.5), 0.3, 0.4, 0.1, "Tetrahedron"),
        (make_octahedron(1.2), 0.4, 0.3, 0.15, "Octahedron"),
        (make_icosahedron(1.3), 0.25, 0.5, 0.0, "Icosahedron"),
        (make_dodecahedron(1.3), 0.35, 0.45, 0.1, "Dodecahedron"),
        (make_torus(1.0, 0.4, 16, 10), 0.8, 0.2, 0.0, "Torus"),
        (make_sphere_wireframe(1.2, 8, 12), 0.6, 0.3, 0.0, "Sphere"),
        (make_stellated_octahedron(0.9), 0.3, 0.5, 0.15, "Stella Octangula"),
        (make_prism(6, 1.0, 1.5), 0.4, 0.35, 0.1, "Hexagonal Prism"),
        (make_prism(5, 1.0, 1.5), 0.35, 0.4, 0.15, "Pentagonal Prism"),
        (make_moebius_strip(24, 0.35, 1.0), 0.5, 0.3, 0.0, "Möbius Strip"),
    ]

    camera = Camera(
        position=(3.5, 2.5, 3.0),
        target=(0, 0, 0),
        up=(0, 0, 1),
        fov=50,
        width=800,
        height=800,
    )
    config = RenderConfig()

    generated = []
    for mesh, rx, ry, rz, label in shapes:
        print(f"  Generating {label}...")
        svg = render_continuous_svg(mesh, camera, config, rx, ry, rz, title=label)
        fname = mesh.name.replace(" ", "_").lower() + ".svg"
        with open(fname, "w") as f:
            f.write(svg)
        generated.append((label, fname, len(mesh.edges), len(mesh.vertices)))
        print(f"    -> {fname} ({len(mesh.edges)} edges, {len(mesh.vertices)} vertices)")

    # Generate combined gallery HTML
    print("  Generating gallery.html...")
    html = generate_gallery_html(generated)
    with open("gallery.html", "w") as f:
        f.write(html)

    return generated


def generate_gallery_html(items):
    """Generate an HTML gallery page embedding all SVGs."""
    cards = ""
    for label, fname, n_edges, n_verts in items:
        cards += f"""
        <div class="card">
            <object data="{fname}" type="image/svg+xml" width="380" height="380"></object>
            <div class="caption">{label} — {n_verts} vertices, {n_edges} edges</div>
        </div>"""

    return f"""<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>3D Pen Plotter — Continuous Stroke Gallery</title>
<style>
  body {{
    font-family: 'IBM Plex Mono', 'Courier New', monospace;
    background: #faf9f6;
    color: #1a1a2e;
    max-width: 1200px;
    margin: 0 auto;
    padding: 2rem;
  }}
  h1 {{ text-align: center; font-weight: 400; margin-bottom: 0.5rem; }}
  .subtitle {{ text-align: center; color: #666; margin-bottom: 2rem; font-size: 0.9rem; }}
  .gallery {{
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
    gap: 1.5rem;
  }}
  .card {{
    background: white;
    border: 1px solid #e0e0e0;
    border-radius: 8px;
    padding: 1rem;
    text-align: center;
    box-shadow: 0 2px 8px rgba(0,0,0,0.05);
  }}
  .card object {{ max-width: 100%; }}
  .caption {{
    margin-top: 0.5rem;
    font-size: 0.85rem;
    color: #555;
  }}
</style>
</head>
<body>
  <h1>3D Pen Plotter — Continuous Stroke</h1>
  <p class="subtitle">Each shape drawn in a single, continuous stroke without lifting the pen. Depth conveyed via line weight and opacity.</p>
  <div class="gallery">{cards}
  </div>
  <footer style="text-align:center; margin-top:2rem; color:#999; font-size:0.8rem;">
    Red dot marks the starting point of each continuous stroke.
  </footer>
</body>
</html>"""


if __name__ == "__main__":
    import os
    os.chdir(os.path.dirname(os.path.abspath(__file__)))
    print("Generating 3D continuous-stroke pen plotter SVGs...")
    results = generate_all()
    print(f"\nDone! Generated {len(results)} SVGs + gallery.html")
