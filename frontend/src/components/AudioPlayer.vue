<template>
  <view class="audio-player" :class="{ compact: compact }">
    <view class="player-controls">
      <!-- 播放/暂停按钮 -->
      <button class="play-btn" @tap="togglePlay" :disabled="!audioUrl || loading">
        <text v-if="loading" class="loading-icon">⏳</text>
        <text v-else-if="isPlaying" class="play-icon">⏸️</text>
        <text v-else class="play-icon">▶️</text>
      </button>

      <!-- 进度信息 -->
      <view class="progress-info">
        <view v-if="!compact" class="time-display">
          <text class="current-time">{{ formatTime(currentTime) }}</text>
          <text class="duration">{{ formatTime(duration) }}</text>
        </view>

        <!-- 进度条 -->
        <view class="progress-bar" @tap="onProgressTap">
          <view class="progress-track">
            <view class="progress-fill" :style="{ width: progressPercent + '%' }"></view>
            <view class="progress-thumb" :style="{ left: progressPercent + '%' }"></view>
          </view>
        </view>

        <!-- 紧凑模式时间显示 -->
        <view v-if="compact" class="compact-time">
          <text>{{ formatTime(currentTime) }} / {{ formatTime(duration) }}</text>
        </view>
      </view>

      <!-- 额外控制按钮 -->
      <view v-if="showControls" class="extra-controls">
        <button class="control-btn" @tap="toggleMute">
          <text>{{ isMuted ? '🔇' : '🔊' }}</text>
        </button>

        <button class="control-btn" @tap="changeSpeed">
          <text class="speed-text">{{ speed }}x</text>
        </button>

        <button v-if="downloadable" class="control-btn" @tap="downloadAudio">
          <text>📥</text>
        </button>
      </view>
    </view>

    <!-- 音频波形显示（可选） -->
    <view v-if="showWaveform && isPlaying" class="waveform">
      <view
        class="wave-bar"
        v-for="i in 30"
        :key="i"
        :style="{ height: getWaveHeight(i) }"
      ></view>
    </view>

    <!-- 错误提示 -->
    <view v-if="error" class="error-message">
      <text class="error-text">{{ error }}</text>
      <button class="retry-btn" @tap="retry">重试</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'

interface Props {
  audioUrl: string
  autoPlay?: boolean
  loop?: boolean
  compact?: boolean
  showControls?: boolean
  showWaveform?: boolean
  downloadable?: boolean
  title?: string
}

interface Emits {
  (e: 'play'): void
  (e: 'pause'): void
  (e: 'ended'): void
  (e: 'error', error: string): void
  (e: 'timeupdate', time: number): void
  (e: 'loadedmetadata', duration: number): void
}

const props = withDefaults(defineProps<Props>(), {
  autoPlay: false,
  loop: false,
  compact: false,
  showControls: true,
  showWaveform: false,
  downloadable: false
})

const emit = defineEmits<Emits>()

const isPlaying = ref<boolean>(false)
const loading = ref<boolean>(false)
const currentTime = ref<number>(0)
const duration = ref<number>(0)
const isMuted = ref<boolean>(false)
const speed = ref<number>(1)
const error = ref<string>('')
const waveHeights = ref<number[]>(Array(30).fill(20))

let audioContext: any = null
let progressTimer: any = null
let waveTimer: any = null

const progressPercent = computed(() => {
  return duration.value > 0 ? (currentTime.value / duration.value) * 100 : 0
})

const speedOptions = [0.5, 0.75, 1, 1.25, 1.5, 2]

watch(() => props.audioUrl, (newUrl) => {
  if (newUrl) {
    loadAudio()
  }
}, { immediate: true })

onMounted(() => {
  initAudioContext()
})

onUnmounted(() => {
  cleanupAudio()
})

const initAudioContext = () => {
  try {
    // #ifdef MP-WEIXIN || APP-PLUS
    audioContext = uni.createInnerAudioContext()

    audioContext.onCanplay(() => {
      loading.value = false
      duration.value = audioContext.duration || 0
      emit('loadedmetadata', duration.value)

      if (props.autoPlay) {
        play()
      }
    })

    audioContext.onPlay(() => {
      isPlaying.value = true
      startProgressTimer()
      if (props.showWaveform) {
        startWaveAnimation()
      }
      emit('play')
    })

    audioContext.onPause(() => {
      isPlaying.value = false
      stopProgressTimer()
      stopWaveAnimation()
      emit('pause')
    })

    audioContext.onStop(() => {
      isPlaying.value = false
      stopProgressTimer()
      stopWaveAnimation()
      currentTime.value = 0
    })

    audioContext.onEnded(() => {
      isPlaying.value = false
      stopProgressTimer()
      stopWaveAnimation()
      currentTime.value = 0
      emit('ended')

      if (props.loop) {
        play()
      }
    })

    audioContext.onError((err: any) => {
      console.error('音频播放错误:', err)
      loading.value = false
      isPlaying.value = false
      error.value = '音频播放失败'
      emit('error', error.value)
    })

    audioContext.onTimeUpdate(() => {
      currentTime.value = audioContext.currentTime || 0
      emit('timeupdate', currentTime.value)
    })
    // #endif

    // #ifdef H5
    initWebAudio()
    // #endif
  } catch (err) {
    console.error('初始化音频上下文失败:', err)
    error.value = '音频初始化失败'
    emit('error', error.value)
  }
}

const initWebAudio = () => {
  // Web Audio API 实现
  console.log('初始化Web音频')
}

const loadAudio = () => {
  if (!audioContext || !props.audioUrl) return

  loading.value = true
  error.value = ''

  try {
    // #ifdef MP-WEIXIN || APP-PLUS
    audioContext.src = props.audioUrl
    // #endif

    // #ifdef H5
    loadWebAudio()
    // #endif
  } catch (err) {
    loading.value = false
    error.value = '音频加载失败'
    emit('error', error.value)
  }
}

const loadWebAudio = () => {
  // Web音频加载
  setTimeout(() => {
    loading.value = false
    duration.value = 30 // 模拟时长
    emit('loadedmetadata', duration.value)
  }, 1000)
}

const play = () => {
  if (!audioContext || loading.value) return

  try {
    // #ifdef MP-WEIXIN || APP-PLUS
    audioContext.play()
    // #endif

    // #ifdef H5
    playWebAudio()
    // #endif
  } catch (err) {
    error.value = '播放失败'
    emit('error', error.value)
  }
}

const pause = () => {
  if (!audioContext) return

  try {
    // #ifdef MP-WEIXIN || APP-PLUS
    audioContext.pause()
    // #endif

    // #ifdef H5
    pauseWebAudio()
    // #endif
  } catch (err) {
    error.value = '暂停失败'
    emit('error', error.value)
  }
}

const stop = () => {
  if (!audioContext) return

  try {
    // #ifdef MP-WEIXIN || APP-PLUS
    audioContext.stop()
    // #endif

    currentTime.value = 0
  } catch (err) {
    console.error('停止音频失败:', err)
  }
}

const togglePlay = () => {
  if (isPlaying.value) {
    pause()
  } else {
    play()
  }
}

const seek = (time: number) => {
  if (!audioContext) return

  time = Math.max(0, Math.min(time, duration.value))

  try {
    // #ifdef MP-WEIXIN || APP-PLUS
    audioContext.seek(time)
    // #endif

    currentTime.value = time
  } catch (err) {
    console.error('跳转失败:', err)
  }
}

const onProgressTap = (e: any) => {
  const { target, detail } = e
  if (!target || duration.value === 0) return

  // 获取点击位置
  uni.createSelectorQuery().selectAll('.progress-track').boundingClientRect((rects: any[]) => {
    if (rects && rects[0]) {
      const rect = rects[0]
      const clickX = detail.x - rect.left
      const percent = clickX / rect.width
      const targetTime = percent * duration.value
      seek(targetTime)
    }
  }).exec()
}

const toggleMute = () => {
  if (!audioContext) return

  isMuted.value = !isMuted.value

  try {
    // #ifdef MP-WEIXIN || APP-PLUS
    audioContext.volume = isMuted.value ? 0 : 1
    // #endif
  } catch (err) {
    console.error('静音切换失败:', err)
  }
}

const changeSpeed = () => {
  const currentIndex = speedOptions.indexOf(speed.value)
  const nextIndex = (currentIndex + 1) % speedOptions.length
  speed.value = speedOptions[nextIndex]

  try {
    // #ifdef MP-WEIXIN || APP-PLUS
    audioContext.playbackRate = speed.value
    // #endif
  } catch (err) {
    console.error('播放速度切换失败:', err)
  }
}

const downloadAudio = () => {
  if (!props.audioUrl) return

  // #ifdef H5
  const link = document.createElement('a')
  link.href = props.audioUrl
  link.download = props.title || `audio_${Date.now()}.mp3`
  link.click()
  // #endif

  // #ifdef MP-WEIXIN || APP-PLUS
  uni.downloadFile({
    url: props.audioUrl,
    success: (res) => {
      uni.showToast({
        title: '下载成功',
        icon: 'success'
      })
    },
    fail: () => {
      uni.showToast({
        title: '下载失败',
        icon: 'none'
      })
    }
  })
  // #endif
}

const retry = () => {
  error.value = ''
  loadAudio()
}

const startProgressTimer = () => {
  progressTimer = setInterval(() => {
    if (audioContext && isPlaying.value) {
      // #ifdef MP-WEIXIN || APP-PLUS
      currentTime.value = audioContext.currentTime || 0
      // #endif

      // #ifdef H5
      currentTime.value += 0.1 // 模拟进度
      // #endif

      emit('timeupdate', currentTime.value)
    }
  }, 100)
}

const stopProgressTimer = () => {
  if (progressTimer) {
    clearInterval(progressTimer)
    progressTimer = null
  }
}

const startWaveAnimation = () => {
  waveTimer = setInterval(() => {
    waveHeights.value = waveHeights.value.map(() => {
      return Math.random() * 30 + 10 // 10-40rpx
    })
  }, 150)
}

const stopWaveAnimation = () => {
  if (waveTimer) {
    clearInterval(waveTimer)
    waveTimer = null
  }
  waveHeights.value = Array(30).fill(20)
}

const getWaveHeight = (index: number): string => {
  return `${waveHeights.value[index - 1] || 20}rpx`
}

const formatTime = (seconds: number): string => {
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
}

const playWebAudio = () => {
  isPlaying.value = true
  startProgressTimer()
  if (props.showWaveform) {
    startWaveAnimation()
  }
  emit('play')
}

const pauseWebAudio = () => {
  isPlaying.value = false
  stopProgressTimer()
  stopWaveAnimation()
  emit('pause')
}

const cleanupAudio = () => {
  stopProgressTimer()
  stopWaveAnimation()

  if (audioContext) {
    try {
      // #ifdef MP-WEIXIN || APP-PLUS
      audioContext.destroy()
      // #endif
    } catch (err) {
      console.error('清理音频上下文失败:', err)
    }
    audioContext = null
  }
}

// 暴露给父组件的方法
defineExpose({
  play,
  pause,
  stop,
  seek,
  togglePlay,
  isPlaying: computed(() => isPlaying.value),
  currentTime: computed(() => currentTime.value),
  duration: computed(() => duration.value),
  progress: computed(() => progressPercent.value)
})
</script>

<style lang="scss" scoped>
.audio-player {
  background: white;
  border-radius: 15rpx;
  padding: 20rpx;
  box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.1);

  &.compact {
    padding: 15rpx;

    .player-controls {
      gap: 15rpx;
    }

    .play-btn {
      width: 60rpx;
      height: 60rpx;
      font-size: 20rpx;
    }
  }
}

.player-controls {
  display: flex;
  align-items: center;
  gap: 20rpx;

  .play-btn {
    width: 80rpx;
    height: 80rpx;
    border-radius: 50%;
    background: linear-gradient(45deg, #007AFF, #5856D6);
    border: none;
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    font-size: 24rpx;
    flex-shrink: 0;

    &:disabled {
      background: #ccc;
    }

    .loading-icon {
      animation: spin 1s linear infinite;
    }
  }

  .progress-info {
    flex: 1;
    min-width: 0;

    .time-display {
      display: flex;
      justify-content: space-between;
      margin-bottom: 10rpx;

      .current-time, .duration {
        font-size: 22rpx;
        color: #666;
      }
    }

    .compact-time {
      margin-top: 8rpx;
      text-align: center;

      text {
        font-size: 20rpx;
        color: #666;
      }
    }
  }

  .extra-controls {
    display: flex;
    gap: 10rpx;
    flex-shrink: 0;

    .control-btn {
      width: 60rpx;
      height: 60rpx;
      border-radius: 50%;
      background: #f8f9fa;
      border: 1rpx solid #dee2e6;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 20rpx;

      .speed-text {
        font-size: 18rpx;
        font-weight: bold;
        color: #007AFF;
      }
    }
  }
}

.progress-bar {
  position: relative;
  padding: 15rpx 0;
  cursor: pointer;

  .progress-track {
    height: 6rpx;
    background: #e9ecef;
    border-radius: 3rpx;
    position: relative;
    overflow: hidden;

    .progress-fill {
      height: 100%;
      background: linear-gradient(45deg, #007AFF, #5856D6);
      border-radius: 3rpx;
      transition: width 0.1s ease;
    }

    .progress-thumb {
      position: absolute;
      top: 50%;
      transform: translate(-50%, -50%);
      width: 16rpx;
      height: 16rpx;
      background: #007AFF;
      border-radius: 50%;
      box-shadow: 0 2rpx 6rpx rgba(0, 122, 255, 0.3);
      transition: left 0.1s ease;
    }
  }
}

.waveform {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 3rpx;
  margin-top: 20rpx;
  height: 60rpx;

  .wave-bar {
    width: 4rpx;
    background: linear-gradient(45deg, #007AFF, #5856D6);
    border-radius: 2rpx;
    transition: height 0.15s ease;
  }
}

.error-message {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff5f5;
  border: 1rpx solid #fed7d7;
  border-radius: 10rpx;
  padding: 15rpx;
  margin-top: 15rpx;

  .error-text {
    color: #e53e3e;
    font-size: 24rpx;
    flex: 1;
  }

  .retry-btn {
    background: #e53e3e;
    color: white;
    border: none;
    border-radius: 8rpx;
    padding: 8rpx 16rpx;
    font-size: 22rpx;
  }
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>