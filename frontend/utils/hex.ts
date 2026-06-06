export function hexKey(q: number, r: number): string {
  return String.fromCharCode(q + 100) + String.fromCharCode(r + 100)
}

export function hexDistance(q1: number, r1: number, q2: number, r2: number): number {
  return (Math.abs(q1 - q2) + Math.abs(q1 + r1 - q2 - r2) + Math.abs(r1 - r2)) / 2
}

export function hexToPixel(q: number, r: number, size: number): { x: number; y: number } {
  const x = size * (3/2 * q)
  const y = size * (Math.sqrt(3)/2 * q + Math.sqrt(3) * r)
  return { x, y }
}

export function hexCorner(cx: number, cy: number, size: number, i: number): { x: number; y: number } {
  const angleDeg = 60 * i - 30
  const angleRad = Math.PI / 180 * angleDeg
  return {
    x: cx + size * Math.cos(angleRad),
    y: cy + size * Math.sin(angleRad)
  }
}

export function hexPath(cx: number, cy: number, size: number): string {
  const points: string[] = []
  for (let i = 0; i < 6; i++) {
    const corner = hexCorner(cx, cy, size, i)
    points.push(`${corner.x},${corner.y}`)
  }
  return points.join(' ')
}

export function axialDirections: [number, number][] = [
  [1, 0], [1, -1], [0, -1],
  [-1, 0], [-1, 1], [0, 1]
]

export function hexNeighbors(q: number, r: number): [number, number][] {
  return axialDirections.map(([dq, dr]) => [q + dq, r + dr])
}

export function generateHexGrid(radius: number): Array<{ q: number; r: number }> {
  const hexes: Array<{ q: number; r: number }> = []
  for (let q = -radius; q <= radius; q++) {
    const r1 = Math.max(-radius, -q - radius)
    const r2 = Math.min(radius, -q + radius)
    for (let r = r1; r <= r2; r++) {
      hexes.push({ q, r })
    }
  }
  return hexes
}
