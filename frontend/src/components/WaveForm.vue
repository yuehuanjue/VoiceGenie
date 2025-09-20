<template>
  <view class="waveform-container" :class="containerClasses">
    <!-- 波形显示区域 -->
    <view class="waveform" :style="waveformStyle">
      <view
        v-for="(bar, index) in waveBars"
        :key="index"
        class="wave-bar"
        :class="getBarClass(index)"
        :style="getBarStyle(index)"
        @tap="onBarTap(index)"
      ></view>

      <!-- 播放进度指示器 -->
      <view
        v-if="showProgress && duration > 0"
        class="progress-indicator"
        :style="progressStyle"
      ></view>

      <!-- 录音实时指示器 -->
      <view
        v-if="isRecording && activeBarIndex >= 0"
        class="recording-indicator"
        :style="{ left: getBarPosition(activeBarIndex) }"
      ></view>
    </view>

    <!-- 时间轴 -->
    <view v-if="showTimeAxis" class="time-axis">
      <text
        v-for="(time, index) in timeMarkers"
        :key="index"
        class="time-marker"
        :style="{ left: getTimePosition(index) }"
      >
        {{ formatTime(time) }}
      </text>
    </view>

    <!-- 控制按钮 -->
    <view v-if="showControls" class="waveform-controls">
      <button class="control-btn" @tap="onPlayPause">
        <text>{{ isPlaying ? '⏸️' : '▶️' }}</text>
      </button>

      <button v-if="zoomable" class="control-btn" @tap="zoomIn">
        <text>🔍+</text>
      </button>

      <button v-if="zoomable" class="control-btn" @tap="zoomOut">
        <text>🔍-</text>
      </button>

      <button v-if="downloadable" class="control-btn" @tap="download">
        <text>📥</text>
      </button>
    </view>

    <!-- 频率/音量显示 -->
    <view v-if="showFrequency" class="frequency-display">
      <text class="frequency-text">{{ currentFrequency }} Hz</text>
      <text class="volume-text">{{ currentVolume }} dB</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'

interface WaveBar {
  height: number
  frequency?: number
  volume?: number
  timestamp?: number
}

interface Props {
  // 数据源
  audioData?: number[]
  audioUrl?: string
  liveData?: boolean

  // 播放控制
  isPlaying?: boolean
  isRecording?: boolean
  currentTime?: number
  duration?: number

  // 显示选项
  showProgress?: boolean
  showTimeAxis?: boolean
  showControls?: boolean
  showFrequency?: boolean

  // 样式配置
  height?: number
  barCount?: number
  barWidth?: number
  barGap?: number
  color?: string
  activeColor?: string
  backgroundColor?: string

  // 功能选项
  interactive?: boolean
  zoomable?: boolean
  downloadable?: boolean

  // 样式主题
  theme?: 'default' | 'compact' | 'spectrum' | 'minimal'
  shape?: 'bar' | 'line' | 'circle'
}

interface Emits {
  (e: 'seek', time: number): void
  (e: 'play'): void
  (e: 'pause'): void
  (e: 'zoom', level: number): void
  (e: 'download'): void
}

const props = withDefaults(defineProps<Props>(), {
  liveData: false,
  isPlaying: false,
  isRecording: false,
  currentTime: 0,
  duration: 0,
  showProgress: true,
  showTimeAxis: false,
  showControls: false,
  showFrequency: false,
  height: 100,
  barCount: 50,
  barWidth: 4,
  barGap: 2,
  color: '#007AFF',
  activeColor: '#FF4757',
  backgroundColor: '#f0f0f0',
  interactive: true,
  zoomable: false,
  downloadable: false,
  theme: 'default',
  shape: 'bar'
})

const emit = defineEmits<Emits>()

const waveBars = ref<WaveBar[]>([])
const zoomLevel = ref<number>(1)
const activeBarIndex = ref<number>(-1)
const currentFrequency = ref<number>(0)
const currentVolume = ref<number>(0)

let animationFrame: number | null = null
let dataUpdateTimer: any = null

const containerClasses = computed(() => {
  return {
    [`theme-${props.theme}`]: true,
    [`shape-${props.shape}`]: true,
    'interactive': props.interactive,
    'recording': props.isRecording,
    'playing': props.isPlaying
  }
})

const waveformStyle = computed(() => {
  return {
    height: `${props.height}rpx`,
    backgroundColor: props.backgroundColor
  }
})

const progressStyle = computed(() => {
  const progress = props.duration > 0 ? (props.currentTime / props.duration) * 100 : 0
  return {
    left: `${progress}%`
  }
})

const timeMarkers = computed(() => {
  if (!props.showTimeAxis || props.duration <= 0) return []

  const markers = []
  const interval = Math.max(1, Math.floor(props.duration / 5)) // 最多5个时间标记

  for (let i = 0; i <= props.duration; i += interval) {
    markers.push(i)
  }

  return markers
})

watch(() => props.audioData, (newData) => {
  if (newData) {
    updateWaveformFromData(newData)
  }
}, { immediate: true })

watch(() => props.audioUrl, (newUrl) => {
  if (newUrl) {
    loadAudioData(newUrl)
  }
})

watch(() => props.liveData, (isLive) => {
  if (isLive) {
    startLiveDataUpdate()
  } else {
    stopLiveDataUpdate()
  }
})

onMounted(() => {
  initializeWaveform()

  if (props.liveData) {
    startLiveDataUpdate()
  }
})

onUnmounted(() => {
  stopLiveDataUpdate()
  if (animationFrame) {
    cancelAnimationFrame(animationFrame)
  }
})

const initializeWaveform = () => {
  // 初始化波形数据
  waveBars.value = Array.from({ length: props.barCount }, (_, index) => ({
    height: Math.random() * 0.5 + 0.1, // 0.1 到 0.6 之间
    frequency: 440 + (index * 20), // 模拟频率
    volume: -60 + (Math.random() * 40), // -60 到 -20 dB
    timestamp: index * (props.duration / props.barCount)
  }))
}

const updateWaveformFromData = (data: number[]) => {
  const barsPerData = Math.max(1, Math.floor(data.length / props.barCount))

  waveBars.value = Array.from({ length: props.barCount }, (_, index) => {
    const dataIndex = index * barsPerData
    const amplitude = Math.abs(data[dataIndex] || 0)

    return {
      height: Math.min(1, amplitude * 2), // 归一化到 0-1
      frequency: 440 + (index * 20),
      volume: 20 * Math.log10(amplitude + 0.001), // 转换为分贝
      timestamp: index * (props.duration / props.barCount)
    }
  })
}

const loadAudioData = async (url: string) => {
  try {
    // 这里可以使用Web Audio API分析音频文件
    // 暂时使用模拟数据
    const mockData = Array.from({ length: 1000 }, () => Math.random() * 2 - 1)
    updateWaveformFromData(mockData)
  } catch (error) {
    console.error('加载音频数据失败:', error)
  }
}

const startLiveDataUpdate = () => {
  dataUpdateTimer = setInterval(() => {
    if (props.isRecording || props.isPlaying) {
      updateLiveData()
    }
  }, 100) // 每100ms更新一次
}

const stopLiveDataUpdate = () => {
  if (dataUpdateTimer) {
    clearInterval(dataUpdateTimer)
    dataUpdateTimer = null
  }
}

const updateLiveData = () => {
  const currentBarIndex = Math.floor(
    (props.currentTime / props.duration) * props.barCount
  )

  activeBarIndex.value = currentBarIndex

  if (props.isRecording) {
    // 模拟录音数据更新
    const amplitude = Math.random() * 0.8 + 0.2
    currentFrequency.value = 440 + Math.random() * 1000
    currentVolume.value = -40 + Math.random() * 30

    if (currentBarIndex < waveBars.value.length) {
      waveBars.value[currentBarIndex] = {
        height: amplitude,
        frequency: currentFrequency.value,
        volume: currentVolume.value,
        timestamp: props.currentTime
      }
    }
  }
}

const getBarClass = (index: number) => {
  return {
    'active': index === activeBarIndex.value,
    'played': props.duration > 0 &&
             (index / props.barCount) <= (props.currentTime / props.duration)
  }
}

const getBarStyle = (index: number) => {
  const bar = waveBars.value[index]
  const height = `${bar.height * props.height * 0.8}rpx`
  const width = `${props.barWidth}rpx`

  let backgroundColor = props.color
  if (index === activeBarIndex.value) {
    backgroundColor = props.activeColor
  }

  return {
    height,
    width,
    backgroundColor,
    marginRight: `${props.barGap}rpx`
  }
}

const getBarPosition = (index: number) => {
  const totalWidth = (props.barWidth + props.barGap) * props.barCount
  return `${(index / props.barCount) * 100}%`
}

const getTimePosition = (index: number) => {
  return `${(index / (timeMarkers.value.length - 1)) * 100}%`
}

const onBarTap = (index: number) => {
  if (!props.interactive || props.duration <= 0) return

  const targetTime = (index / props.barCount) * props.duration
  emit('seek', targetTime)
}

const onPlayPause = () => {
  if (props.isPlaying) {
    emit('pause')
  } else {
    emit('play')
  }
}

const zoomIn = () => {
  zoomLevel.value = Math.min(zoomLevel.value * 1.5, 5)
  emit('zoom', zoomLevel.value)
}

const zoomOut = () => {
  zoomLevel.value = Math.max(zoomLevel.value / 1.5, 0.5)
  emit('zoom', zoomLevel.value)
}

const download = () => {
  emit('download')
}

const formatTime = (seconds: number): string => {
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

// 暴露给父组件的方法
defineExpose({
  updateData: updateWaveformFromData,
  seek: (time: number) => {
    activeBarIndex.value = Math.floor((time / props.duration) * props.barCount)
  },
  zoomLevel: computed(() => zoomLevel.value)
})
</script>

<style lang="scss" scoped>
.waveform-container {
  width: 100%;
  position: relative;

  &.interactive {
    cursor: pointer;
  }

  &.recording .waveform {
    border: 2rpx solid #FF4757;
    animation: recording-glow 2s infinite;
  }

  &.playing .progress-indicator {
    animation: pulse 1s infinite;
  }
}

.waveform {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20rpx;
  border-radius: 10rpx;
  position: relative;
  overflow: hidden;

  .wave-bar {
    border-radius: 2rpx;
    transition: all 0.2s ease;
    transform-origin: bottom;

    &.active {
      transform: scaleY(1.2);
      filter: brightness(1.3);
    }

    &.played {
      opacity: 0.7;
    }

    // 不同形状的波形条
    .shape-line & {
      border-radius: 0;
      height: 2rpx !important;
      transform-origin: center;
    }

    .shape-circle & {
      border-radius: 50%;
      width: 8rpx !important;
    }
  }

  .progress-indicator {
    position: absolute;
    top: 0;
    bottom: 0;
    width: 2rpx;
    background: #FF4757;
    z-index: 10;
    pointer-events: none;
  }

  .recording-indicator {
    position: absolute;
    top: 50%;
    transform: translateY(-50%);
    width: 4rpx;
    height: 80%;
    background: #FF4757;
    border-radius: 2rpx;
    z-index: 11;
    animation: recording-pulse 0.5s infinite alternate;
  }
}

.time-axis {
  position: relative;
  height: 40rpx;
  margin-top: 10rpx;

  .time-marker {
    position: absolute;
    font-size: 20rpx;
    color: #666;
    transform: translateX(-50%);
  }
}

.waveform-controls {
  display: flex;
  justify-content: center;
  gap: 20rpx;
  margin-top: 20rpx;

  .control-btn {
    width: 60rpx;
    height: 60rpx;
    border-radius: 50%;
    background: #f8f9fa;
    border: 1rpx solid #dee2e6;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 24rpx;
  }
}

.frequency-display {
  display: flex;
  justify-content: space-between;
  margin-top: 15rpx;
  padding: 0 20rpx;

  .frequency-text, .volume-text {
    font-size: 22rpx;
    color: #666;
  }
}

// 主题样式
.theme-compact {
  .waveform {
    padding: 10rpx;
    height: 60rpx !important;
  }

  .wave-bar {
    width: 2rpx !important;
    margin-right: 1rpx !important;
  }
}

.theme-spectrum {
  .waveform {
    background: linear-gradient(45deg, #667eea, #764ba2);
  }

  .wave-bar {
    background: rgba(255, 255, 255, 0.8) !important;
    box-shadow: 0 0 10rpx rgba(255, 255, 255, 0.5);
  }
}

.theme-minimal {
  .waveform {
    background: transparent;
    border: 1rpx solid #eee;
  }

  .wave-bar {
    background: #ddd !important;

    &.active {
      background: #007AFF !important;
    }
  }
}

@keyframes recording-glow {
  0%, 100% {
    border-color: #FF4757;
    box-shadow: 0 0 10rpx rgba(255, 71, 87, 0.3);
  }
  50% {
    border-color: #FF6B6B;
    box-shadow: 0 0 20rpx rgba(255, 107, 107, 0.6);
  }
}

@keyframes recording-pulse {
  0% {
    opacity: 0.5;
    transform: translateY(-50%) scaleY(0.8);
  }
  100% {
    opacity: 1;
    transform: translateY(-50%) scaleY(1.2);
  }
}

@keyframes pulse {
  0%, 100% {
    transform: scaleX(1);
    opacity: 1;
  }
  50% {
    transform: scaleX(1.2);
    opacity: 0.8;
  }
}
</style>