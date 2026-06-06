<template>
  <div class="hex-map-container" ref="containerRef">
    <svg :width="svgWidth" :height="svgHeight" class="hex-map-svg">
      <defs>
        <filter id="glow">
          <feGaussianBlur stdDeviation="2" result="coloredBlur"/>
          <feMerge>
            <feMergeNode in="coloredBlur"/>
            <feMergeNode in="SourceGraphic"/>
          </feMerge>
        </filter>
      </defs>
      
      <g :transform="`translate(${offsetX}, ${offsetY})`">
        <g v-for="hex in hexList" :key="`hex-${hex.q}-${hex.r}`">
          <polygon
            :points="getHexPath(hex.q, hex.r)"
            :fill="getHexColor(hex.q, hex.r)"
            :stroke="getHexStroke(hex.q, hex.r)"
            :stroke-width="isSelected(hex.q, hex.r) ? 3 : 1"
            class="hex-tile"
            :class="{ 'selected': isSelected(hex.q, hex.r), 'clickable': canClick(hex.q, hex.r) }"
            @click="handleHexClick(hex.q, hex.r)"
          />
          
          <g v-if="hexData(hex.q, hex.r)?.facility" @click.stop="handleFacilityClick(hex.q, hex.r)">
            <text
              :x="hexPixel(hex.q, hex.r).x"
              :y="hexPixel(hex.q, hex.r).y - 8"
              text-anchor="middle"
              class="facility-icon"
            >
              {{ getFacilityIcon(hexData(hex.q, hex.r)?.facility?.type) }}
            </text>
          </g>
          
          <g v-if="getShipsAt(hex.q, hex.r).length > 0">
            <text
              :x="hexPixel(hex.q, hex.r).x"
              :y="hexPixel(hex.q, hex.r).y + 12"
              text-anchor="middle"
              class="ship-icon"
            >
              {{ getShipIcon(getShipsAt(hex.q, hex.r)[0]?.type) }}
            </text>
            <text
              v-if="getShipsAt(hex.q, hex.r).length > 1"
              :x="hexPixel(hex.q, hex.r).x + 10"
              :y="hexPixel(hex.q, hex.r).y + 5"
              text-anchor="start"
              class="ship-count"
            >
              ×{{ getShipsAt(hex.q, hex.r).length }}
            </text>
          </g>
          
          <g v-if="hexData(hex.q, hex.r)?.has_current">
            <text
              :x="hexPixel(hex.q, hex.r).x - 12"
              :y="hexPixel(hex.q, hex.r).y - 12"
              class="current-indicator"
            >
              ➤
            </text>
          </g>
        </g>
        
        <g v-for="typhoon in typhoons" :key="`typhoon-${typhoon.id}`">
          <circle
            :cx="hexPixel(typhoon.hex_q, typhoon.hex_r).x"
            :cy="hexPixel(typhoon.hex_q, typhoon.hex_r).y"
            :r="hexSize * 0.8"
            fill="rgba(231, 76, 60, 0.3)"
            stroke="#E74C3C"
            stroke-width="2"
          />
          <text
            :x="hexPixel(typhoon.hex_q, typhoon.hex_r).x"
            :y="hexPixel(typhoon.hex_q, typhoon.hex_r).y + 5"
            text-anchor="middle"
            class="typhoon-icon"
          >
            🌀
          </text>
        </g>
      </g>
    </svg>
  </div>
</template>

<script setup lang="ts">
import type { Ship, Facility } from '~/types'
import { hexKey, hexToPixel, hexPath } from '~/utils/hex'

interface Props {
  radius: number
  hexSize?: number
}

const props = withDefaults(defineProps<Props>(), {
  hexSize: 40
})

const emit = defineEmits<{
  (e: 'hexClick', q: number, r: number): void
  (e: 'facilityClick', q: number, r: number): void
}>()

const containerRef = ref<HTMLDivElement | null>(null)
const gameStore = useGameStore()

const hexList = computed(() => {
  if (!gameStore.gameState) return []
  return Object.values(gameStore.gameState.hexes).map(h => ({ q: h.q, r: h.r }))
})

const svgWidth = computed(() => props.radius * props.hexSize * 3.5 + 100)
const svgHeight = computed(() => props.radius * props.hexSize * 3.5 + 100)
const offsetX = computed(() => svgWidth.value / 2)
const offsetY = computed(() => svgHeight.value / 2)

const typhoons = computed(() => gameStore.gameState?.typhoons || [])

function hexPixel(q: number, r: number) {
  return hexToPixel(q, r, props.hexSize)
}

function getHexPath(q: number, r: number): string {
  const pixel = hexPixel(q, r)
  return hexPath(pixel.x, pixel.y, props.hexSize * 0.95)
}

function hexData(q: number, r: number) {
  return gameStore.getHexAt(q, r)
}

function getHexColor(q: number, r: number): string {
  const hex = hexData(q, r)
  if (!hex) return '#ccc'
  if (!hex.discovered) return '#1a1a2e'
  
  const TERRAIN_COLORS: Record<string, string> = {
    shallow: '#5DADE2',
    deep: '#2874A6',
    trench: '#1B4F72',
    reef: '#58D68D',
    vent: '#E67E22',
    open_ocean: '#3498DB'
  }
  
  let color = TERRAIN_COLORS[hex.terrain] || '#3498DB'
  
  if (hex.pollution > 50) {
    color = '#566573'
  }
  
  if (hex.owner_id && gameStore.currentPlayerId && hex.owner_id === gameStore.currentPlayerId) {
    return lightenColor(color, 20)
  }
  
  return color
}

function getHexStroke(q: number, r: number): string {
  const hex = hexData(q, r)
  if (!hex) return '#fff'
  
  if (hex.is_eez && hex.eez_owner_id) {
    const player = gameStore.gameState?.players[hex.eez_owner_id]
    return player?.color || '#fff'
  }
  
  return 'rgba(255,255,255,0.3)'
}

function isSelected(q: number, r: number): boolean {
  return gameStore.selectedHex?.q === q && gameStore.selectedHex?.r === r
}

function canClick(q: number, r: number): boolean {
  const hex = hexData(q, r)
  return hex?.discovered || false
}

function getShipsAt(q: number, r: number): Ship[] {
  if (!gameStore.gameState) return []
  return gameStore.gameState.ships.filter(s => s.hex_q === q && s.hex_r === r)
}

function getFacilityIcon(type?: string): string {
  const icons: Record<string, string> = {
    drilling: '🏗️',
    mine: '⛏️',
    tidal: '⚡',
    farm: '🐟',
    port: '⚓'
  }
  return icons[type || ''] || '🏗️'
}

function getShipIcon(type?: string): string {
  const icons: Record<string, string> = {
    explorer: '🔍',
    constructor: '🔧',
    transport: '📦',
    escort: '⚔️'
  }
  return icons[type || ''] || '🚢'
}

function lightenColor(color: string, percent: number): string {
  const num = parseInt(color.replace('#', ''), 16)
  const amt = Math.round(2.55 * percent)
  const R = (num >> 16) + amt
  const G = (num >> 8 & 0x00FF) + amt
  const B = (num & 0x0000FF) + amt
  return '#' + (
    0x1000000 +
    (R < 255 ? R < 1 ? 0 : R : 255) * 0x10000 +
    (G < 255 ? G < 1 ? 0 : G : 255) * 0x100 +
    (B < 255 ? B < 1 ? 0 : B : 255)
  ).toString(16).slice(1)
}

function handleHexClick(q: number, r: number) {
  if (!canClick(q, r)) return
  gameStore.selectHex(q, r)
  emit('hexClick', q, r)
}

function handleFacilityClick(q: number, r: number) {
  const hex = hexData(q, r)
  if (hex?.facility) {
    gameStore.selectFacility(hex.facility)
    emit('facilityClick', q, r)
  }
}
</script>

<style scoped>
.hex-map-container {
  width: 100%;
  height: 100%;
  overflow: auto;
  background: linear-gradient(135deg, #0c1929 0%, #1a3a5c 100%);
  border-radius: 8px;
}

.hex-map-svg {
  display: block;
  margin: 0 auto;
}

.hex-tile {
  cursor: pointer;
  transition: all 0.2s ease;
}

.hex-tile:hover {
  filter: brightness(1.2);
}

.hex-tile.selected {
  filter: brightness(1.3);
  stroke: #f1c40f !important;
}

.hex-tile.clickable {
  cursor: pointer;
}

.facility-icon {
  font-size: 18px;
  pointer-events: none;
}

.ship-icon {
  font-size: 16px;
  pointer-events: none;
}

.ship-count {
  font-size: 11px;
  fill: #fff;
  font-weight: bold;
  pointer-events: none;
}

.current-indicator {
  font-size: 14px;
  fill: #3498db;
  pointer-events: none;
}

.typhoon-icon {
  font-size: 20px;
  animation: spin 2s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
