<template>
  <view class="audio-recorder">
    <!-- 录音按钮 -->
    <view class="record-button-container">
      <button
        class="record-button"
        :class="{ 'recording': isRecording, 'disabled': disabled }"
        @touchstart="startRecording"
        @touchend="stopRecording"
        @touchcancel="cancelRecording"
        :disabled="disabled"
      >
        <view class="record-icon">
          <text v-if="!isRecording">🎤</text>
          <view v-else class="recording-animation">
            <view class="pulse"></view>
            <text>🎤</text>
          </view>
        </view>
        <text class="record-text">
          {{ getRecordText() }}
        </text>
      </button>
    </view>

    <!-- 录音状态信息 -->
    <view v-if="isRecording" class="recording-info">
      <text class="recording-time">{{ formatTime(recordingDuration) }}</text>
      <view class="recording-wave">
        <view
          class="wave-bar"
          v-for="i in 20"
          :key="i"
          :style="{ height: getWaveHeight(i) }"
        ></view>
      </view>
      <text class="recording-tip">{{ recordingTip }}</text>
    </view>

    <!-- 录音控制按钮 -->
    <view v-if="showControls && isRecording" class="control-buttons">
      <button class="control-btn cancel" @tap="cancelRecording">
        <text class="control-icon">❌</text>
        <text class="control-text">取消</text>
      </button>
      <button class="control-btn send" @tap="stopRecording">
        <text class="control-icon">✅</text>
        <text class="control-text">发送</text>
      </button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

interface RecordResult {
  tempFilePath: string
  duration: number
  fileSize?: number
}

interface Props {
  disabled?: boolean
  maxDuration?: number
  minDuration?: number
  showControls?: boolean
  autoSend?: boolean
  quality?: 'low' | 'standard' | 'high'
}

interface Emits {
  (e: 'start'): void
  (e: 'stop', result: RecordResult): void
  (e: 'cancel'): void
  (e: 'error', error: any): void
  (e: 'progress', duration: number): void
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  maxDuration: 60,
  minDuration: 1,
  showControls: true,
  autoSend: true,
  quality: 'standard'
})

const emit = defineEmits<Emits>()

const isRecording = ref<boolean>(false)
const recordingDuration = ref<number>(0)
const waveHeights = ref<number[]>(Array(20).fill(20))

let recorderManager: any = null
let recordingTimer: any = null
let waveTimer: any = null

const recordingTip = computed(() => {
  if (recordingDuration.value < props.minDuration) {
    return `至少录制 ${props.minDuration} 秒`
  }
  if (recordingDuration.value >= props.maxDuration - 5) {
    return `即将达到最大时长 ${props.maxDuration} 秒`
  }
  return '松开发送，继续录制'
})

onMounted(() => {
  initRecorder()
})

onUnmounted(() => {
  cleanupRecorder()
})

const initRecorder = () => {
  try {
    // #ifdef MP-WEIXIN || APP-PLUS
    recorderManager = uni.getRecorderManager()

    recorderManager.onStart(() => {
      console.log('录音开始')
      isRecording.value = true
      startTimer()
      startWaveAnimation()
      emit('start')
    })

    recorderManager.onStop((res: any) => {
      console.log('录音结束', res)
      isRecording.value = false
      stopTimer()
      stopWaveAnimation()

      if (res.tempFilePath) {
        const result: RecordResult = {
          tempFilePath: res.tempFilePath,
          duration: res.duration || recordingDuration.value * 1000,
          fileSize: res.fileSize
        }
        emit('stop', result)
      }
    })

    recorderManager.onError((err: any) => {
      console.error('录音错误', err)
      isRecording.value = false
      stopTimer()
      stopWaveAnimation()
      emit('error', err)

      uni.showToast({
        title: '录音失败',
        icon: 'none'
      })
    })
    // #endif

    // #ifdef H5
    // H5环境下使用 MediaRecorder API
    initWebRecorder()
    // #endif
  } catch (error) {
    console.error('初始化录音器失败:', error)
    emit('error', error)
  }
}

const initWebRecorder = () => {
  // Web录音器初始化逻辑
  console.log('初始化Web录音器')
}

const getRecordText = (): string => {
  if (props.disabled) return '录音不可用'
  if (isRecording.value) return props.autoSend ? '松开发送' : '正在录音'
  return '按住说话'
}

const startRecording = () => {
  if (props.disabled || isRecording.value) return

  try {
    recordingDuration.value = 0

    // #ifdef MP-WEIXIN || APP-PLUS
    const recordConfig = getRecordConfig()
    recorderManager.start(recordConfig)
    // #endif

    // #ifdef H5
    startWebRecording()
    // #endif
  } catch (error) {
    emit('error', error)
  }
}

const stopRecording = () => {
  if (!isRecording.value) return

  try {
    // 检查最小录音时长
    if (recordingDuration.value < props.minDuration) {
      uni.showToast({
        title: `录音时长不能少于${props.minDuration}秒`,
        icon: 'none'
      })
      return
    }

    // #ifdef MP-WEIXIN || APP-PLUS
    recorderManager.stop()
    // #endif

    // #ifdef H5
    stopWebRecording()
    // #endif
  } catch (error) {
    emit('error', error)
  }
}

const cancelRecording = () => {
  if (!isRecording.value) return

  try {
    isRecording.value = false
    stopTimer()
    stopWaveAnimation()

    // #ifdef MP-WEIXIN || APP-PLUS
    recorderManager.stop()
    // #endif

    emit('cancel')

    uni.showToast({
      title: '录音已取消',
      icon: 'none'
    })
  } catch (error) {
    emit('error', error)
  }
}

const getRecordConfig = () => {
  const qualityMap = {
    low: {
      sampleRate: 8000,
      encodeBitRate: 32000,
      format: 'mp3'
    },
    standard: {
      sampleRate: 16000,
      encodeBitRate: 96000,
      format: 'mp3'
    },
    high: {
      sampleRate: 44100,
      encodeBitRate: 192000,
      format: 'mp3'
    }
  }

  return {
    duration: props.maxDuration * 1000,
    numberOfChannels: 1,
    ...qualityMap[props.quality]
  }
}

const startTimer = () => {
  recordingTimer = setInterval(() => {
    recordingDuration.value += 1
    emit('progress', recordingDuration.value)

    // 达到最大时长自动停止
    if (recordingDuration.value >= props.maxDuration) {
      stopRecording()
    }
  }, 1000)
}

const stopTimer = () => {
  if (recordingTimer) {
    clearInterval(recordingTimer)
    recordingTimer = null
  }
}

const startWaveAnimation = () => {
  waveTimer = setInterval(() => {
    waveHeights.value = waveHeights.value.map(() => {
      return Math.random() * 40 + 10 // 10-50rpx
    })
  }, 100)
}

const stopWaveAnimation = () => {
  if (waveTimer) {
    clearInterval(waveTimer)
    waveTimer = null
  }
  waveHeights.value = Array(20).fill(20)
}

const getWaveHeight = (index: number): string => {
  return `${waveHeights.value[index - 1] || 20}rpx`
}

const formatTime = (seconds: number): string => {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
}

const startWebRecording = () => {
  // Web录音实现
  console.log('开始Web录音')
  isRecording.value = true
  startTimer()
  startWaveAnimation()
  emit('start')
}

const stopWebRecording = () => {
  // Web录音停止
  console.log('停止Web录音')
  isRecording.value = false
  stopTimer()
  stopWaveAnimation()

  // 模拟录音结果
  const result: RecordResult = {
    tempFilePath: 'blob:' + Date.now(),
    duration: recordingDuration.value * 1000
  }
  emit('stop', result)
}

const cleanupRecorder = () => {
  stopTimer()
  stopWaveAnimation()

  if (recorderManager) {
    recorderManager = null
  }
}

// 暴露给父组件的方法
defineExpose({
  startRecording,
  stopRecording,
  cancelRecording,
  isRecording: computed(() => isRecording.value),
  duration: computed(() => recordingDuration.value)
})
</script>

<style lang="scss" scoped>
.audio-recorder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20rpx;
}

.record-button-container {
  display: flex;
  justify-content: center;
}

.record-button {
  width: 200rpx;
  height: 200rpx;
  border-radius: 50%;
  background: linear-gradient(135deg, #FF6B6B, #FF8E8E);
  border: none;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
  box-shadow: 0 8rpx 24rpx rgba(255, 107, 107, 0.3);

  &.recording {
    transform: scale(1.1);
    background: linear-gradient(135deg, #FF4757, #FF3838);
    box-shadow: 0 12rpx 32rpx rgba(255, 71, 87, 0.4);
  }

  &.disabled {
    background: #ccc;
    box-shadow: none;
  }

  .record-icon {
    font-size: 48rpx;
    margin-bottom: 10rpx;
    position: relative;
  }

  .record-text {
    font-size: 24rpx;
    color: white;
    font-weight: bold;
  }
}

.recording-animation {
  position: relative;

  .pulse {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 80rpx;
    height: 80rpx;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.3);
    animation: pulse 1s infinite;
  }
}

@keyframes pulse {
  0% {
    transform: translate(-50%, -50%) scale(1);
    opacity: 1;
  }
  100% {
    transform: translate(-50%, -50%) scale(1.5);
    opacity: 0;
  }
}

.recording-info {
  text-align: center;
  padding: 20rpx;
  background: rgba(255, 255, 255, 0.9);
  border-radius: 15rpx;
  min-width: 300rpx;

  .recording-time {
    font-size: 32rpx;
    font-weight: bold;
    color: #FF4757;
    margin-bottom: 15rpx;
    display: block;
  }

  .recording-wave {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 4rpx;
    margin-bottom: 15rpx;

    .wave-bar {
      width: 6rpx;
      background: #FF4757;
      border-radius: 3rpx;
      transition: height 0.1s ease;
    }
  }

  .recording-tip {
    font-size: 24rpx;
    color: #666;
    display: block;
  }
}

.control-buttons {
  display: flex;
  gap: 40rpx;

  .control-btn {
    width: 120rpx;
    height: 120rpx;
    border-radius: 50%;
    border: none;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8rpx;

    &.cancel {
      background: #ff4757;
    }

    &.send {
      background: #2ed573;
    }

    .control-icon {
      font-size: 32rpx;
    }

    .control-text {
      font-size: 20rpx;
      color: white;
      font-weight: bold;
    }
  }
}
</style>