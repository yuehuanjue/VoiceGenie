<template>
  <view class="chat-container">
    <!-- 聊天消息区域 -->
    <scroll-view
      class="message-list"
      scroll-y="true"
      :scroll-top="scrollTop"
      scroll-with-animation="true"
    >
      <view class="message-item" v-for="message in messages" :key="message.id">
        <view
          class="message-bubble"
          :class="{ 'user': message.type === 'user', 'ai': message.type === 'ai' }"
        >
          <view class="message-content">
            <text class="message-text">{{ message.text }}</text>
            <view class="message-time">{{ formatTime(message.timestamp) }}</view>
          </view>

          <!-- 音频播放按钮 -->
          <view
            v-if="message.audioUrl"
            class="audio-controls"
            @tap="toggleAudio(message)"
          >
            <text class="audio-icon">
              {{ message.isPlaying ? '⏸️' : '▶️' }}
            </text>
            <text class="audio-duration">{{ message.duration || '0:00' }}</text>
          </view>
        </view>
      </view>

      <!-- 加载中提示 -->
      <view v-if="isProcessing" class="loading-message">
        <view class="message-bubble ai">
          <view class="loading-dots">
            <text class="dot"></text>
            <text class="dot"></text>
            <text class="dot"></text>
          </view>
        </view>
      </view>
    </scroll-view>

    <!-- 录音控制区域 -->
    <view class="record-area">
      <!-- 录音按钮 -->
      <view class="record-button-container">
        <button
          class="record-button"
          :class="{ 'recording': isRecording }"
          @touchstart="startRecording"
          @touchend="stopRecording"
          @touchcancel="cancelRecording"
        >
          <view class="record-icon">
            <text v-if="!isRecording">🎤</text>
            <view v-else class="recording-animation">
              <view class="pulse"></view>
              <text>🎤</text>
            </view>
          </view>
          <text class="record-text">
            {{ isRecording ? '松开发送' : '按住说话' }}
          </text>
        </button>
      </view>

      <!-- 录音时长显示 -->
      <view v-if="isRecording" class="recording-info">
        <text class="recording-time">{{ recordingTime }}s</text>
        <view class="recording-wave">
          <view class="wave-bar" v-for="i in 20" :key="i"></view>
        </view>
      </view>

      <!-- 功能按钮 -->
      <view class="action-buttons">
        <button class="action-btn" @tap="clearMessages">
          <text class="btn-icon">🗑️</text>
          <text class="btn-text">清空</text>
        </button>

        <button class="action-btn" @tap="showSettings">
          <text class="btn-icon">⚙️</text>
          <text class="btn-text">设置</text>
        </button>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, nextTick, onMounted, onUnmounted } from 'vue'

// 消息接口定义
interface Message {
  id: string
  type: 'user' | 'ai'
  text: string
  timestamp: number
  audioUrl?: string
  duration?: string
  isPlaying?: boolean
}

// 响应式数据
const messages = ref<Message[]>([])
const isRecording = ref<boolean>(false)
const isProcessing = ref<boolean>(false)
const scrollTop = ref<number>(0)
const recordingTime = ref<number>(0)

// 录音相关
let recorderManager: any = null
let audioContext: any = null
let recordingTimer: any = null

// 生命周期
onMounted(() => {
  initRecorder()
  initAudioContext()
  loadHistoryMessages()
})

onUnmounted(() => {
  cleanupRecorder()
  cleanupAudioContext()
})

// 初始化录音管理器
const initRecorder = () => {
  // #ifdef MP-WEIXIN || APP-PLUS
  recorderManager = uni.getRecorderManager()

  recorderManager.onStart(() => {
    console.log('录音开始')
    startRecordingTimer()
  })

  recorderManager.onStop((res: any) => {
    console.log('录音结束', res)
    stopRecordingTimer()
    handleRecordingResult(res)
  })

  recorderManager.onError((err: any) => {
    console.error('录音错误', err)
    uni.showToast({
      title: '录音失败',
      icon: 'none'
    })
    resetRecordingState()
  })
  // #endif
}

// 初始化音频上下文
const initAudioContext = () => {
  // #ifdef MP-WEIXIN || APP-PLUS
  audioContext = uni.createInnerAudioContext()

  audioContext.onEnded(() => {
    // 音频播放结束
    messages.value.forEach(msg => {
      if (msg.isPlaying) {
        msg.isPlaying = false
      }
    })
  })

  audioContext.onError((err: any) => {
    console.error('音频播放错误', err)
    uni.showToast({
      title: '音频播放失败',
      icon: 'none'
    })
  })
  // #endif
}

// 开始录音
const startRecording = () => {
  if (isRecording.value) return

  isRecording.value = true
  recordingTime.value = 0

  // #ifdef MP-WEIXIN || APP-PLUS
  recorderManager.start({
    duration: 60000, // 最长60秒
    sampleRate: 16000,
    numberOfChannels: 1,
    encodeBitRate: 96000,
    format: 'mp3'
  })
  // #endif

  // #ifdef H5
  // H5环境下的录音实现
  startWebRecording()
  // #endif
}

// 停止录音
const stopRecording = () => {
  if (!isRecording.value) return

  // #ifdef MP-WEIXIN || APP-PLUS
  recorderManager.stop()
  // #endif

  // #ifdef H5
  stopWebRecording()
  // #endif

  isRecording.value = false
}

// 取消录音
const cancelRecording = () => {
  if (!isRecording.value) return

  isRecording.value = false
  stopRecordingTimer()

  uni.showToast({
    title: '录音已取消',
    icon: 'none'
  })
}

// 处理录音结果
const handleRecordingResult = async (result: any) => {
  if (!result.tempFilePath) {
    uni.showToast({
      title: '录音失败',
      icon: 'none'
    })
    return
  }

  // 添加用户消息
  const userMessage: Message = {
    id: generateId(),
    type: 'user',
    text: '语音消息',
    timestamp: Date.now(),
    audioUrl: result.tempFilePath,
    duration: formatDuration(result.duration || 0)
  }

  messages.value.push(userMessage)
  scrollToBottom()

  // 开始处理AI响应
  isProcessing.value = true

  try {
    // 这里会调用后端API进行语音识别和AI对话
    // const response = await api.processVoiceMessage(result.tempFilePath)

    // 模拟AI响应
    setTimeout(() => {
      const aiMessage: Message = {
        id: generateId(),
        type: 'ai',
        text: '这是AI的回复消息。我听到了您的语音输入，正在为您处理...',
        timestamp: Date.now()
      }

      messages.value.push(aiMessage)
      isProcessing.value = false
      scrollToBottom()
    }, 2000)

  } catch (error) {
    console.error('处理语音消息失败:', error)
    isProcessing.value = false
    uni.showToast({
      title: '处理失败，请重试',
      icon: 'none'
    })
  }
}

// 录音计时器
const startRecordingTimer = () => {
  recordingTimer = setInterval(() => {
    recordingTime.value += 1
    if (recordingTime.value >= 60) {
      stopRecording()
    }
  }, 1000)
}

const stopRecordingTimer = () => {
  if (recordingTimer) {
    clearInterval(recordingTimer)
    recordingTimer = null
  }
}

const resetRecordingState = () => {
  isRecording.value = false
  recordingTime.value = 0
  stopRecordingTimer()
}

// 音频播放控制
const toggleAudio = (message: Message) => {
  if (!message.audioUrl) return

  if (message.isPlaying) {
    // 停止播放
    audioContext.stop()
    message.isPlaying = false
  } else {
    // 停止其他正在播放的音频
    messages.value.forEach(msg => {
      if (msg.isPlaying && msg.id !== message.id) {
        msg.isPlaying = false
      }
    })

    // 开始播放
    audioContext.src = message.audioUrl
    audioContext.play()
    message.isPlaying = true
  }
}

// 清空消息
const clearMessages = () => {
  uni.showModal({
    title: '确认清空',
    content: '确定要清空所有对话记录吗？',
    success: (res) => {
      if (res.confirm) {
        messages.value = []
      }
    }
  })
}

// 显示设置
const showSettings = () => {
  uni.navigateTo({
    url: '/pages/settings/settings'
  })
}

// 滚动到底部
const scrollToBottom = () => {
  nextTick(() => {
    scrollTop.value = 99999
  })
}

// 加载历史消息
const loadHistoryMessages = () => {
  // 这里会从本地存储或API加载历史消息
  const sampleMessages: Message[] = [
    {
      id: generateId(),
      type: 'ai',
      text: '您好！我是VoiceGenie，您的智能语音助手。有什么可以帮助您的吗？',
      timestamp: Date.now() - 60000
    }
  ]

  messages.value = sampleMessages
  scrollToBottom()
}

// H5录音实现（占位）
const startWebRecording = () => {
  console.log('H5录音开始')
  startRecordingTimer()
}

const stopWebRecording = () => {
  console.log('H5录音结束')
  stopRecordingTimer()

  // 模拟录音结果
  handleRecordingResult({
    tempFilePath: 'mock-audio-url',
    duration: recordingTime.value * 1000
  })
}

// 清理资源
const cleanupRecorder = () => {
  if (recorderManager) {
    recorderManager = null
  }
  stopRecordingTimer()
}

const cleanupAudioContext = () => {
  if (audioContext) {
    audioContext.destroy()
    audioContext = null
  }
}

// 工具函数
const generateId = (): string => {
  return Date.now().toString(36) + Math.random().toString(36).substr(2)
}

const formatTime = (timestamp: number): string => {
  const date = new Date(timestamp)
  const hours = date.getHours().toString().padStart(2, '0')
  const minutes = date.getMinutes().toString().padStart(2, '0')
  return `${hours}:${minutes}`
}

const formatDuration = (duration: number): string => {
  const seconds = Math.floor(duration / 1000)
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = seconds % 60
  return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`
}
</script>

<style lang="scss" scoped>
.chat-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: #f5f5f5;
}

.message-list {
  flex: 1;
  padding: 20rpx;
}

.message-item {
  margin-bottom: 30rpx;
}

.message-bubble {
  max-width: 70%;
  padding: 20rpx 30rpx;
  border-radius: 20rpx;

  &.user {
    background: linear-gradient(135deg, #007AFF, #5856D6);
    color: white;
    margin-left: auto;
    border-bottom-right-radius: 8rpx;
  }

  &.ai {
    background: white;
    color: #333;
    margin-right: auto;
    border-bottom-left-radius: 8rpx;
    box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.1);
  }
}

.message-content {
  .message-text {
    display: block;
    line-height: 1.4;
    margin-bottom: 10rpx;
  }

  .message-time {
    font-size: 22rpx;
    opacity: 0.7;
  }
}

.audio-controls {
  display: flex;
  align-items: center;
  gap: 10rpx;
  margin-top: 15rpx;
  padding-top: 15rpx;
  border-top: 1rpx solid rgba(255, 255, 255, 0.2);

  .audio-icon {
    font-size: 24rpx;
  }

  .audio-duration {
    font-size: 22rpx;
    opacity: 0.8;
  }
}

.loading-message {
  display: flex;
  justify-content: flex-start;
  margin-bottom: 30rpx;
}

.loading-dots {
  display: flex;
  gap: 8rpx;

  .dot {
    width: 8rpx;
    height: 8rpx;
    border-radius: 50%;
    background-color: #ccc;
    animation: loading-pulse 1.4s infinite ease-in-out;

    &:nth-child(1) { animation-delay: -0.32s; }
    &:nth-child(2) { animation-delay: -0.16s; }
    &:nth-child(3) { animation-delay: 0s; }
  }
}

@keyframes loading-pulse {
  0%, 80%, 100% {
    transform: scale(0);
    opacity: 0.5;
  }
  40% {
    transform: scale(1);
    opacity: 1;
  }
}

.record-area {
  background: white;
  padding: 30rpx;
  border-top: 1rpx solid #eee;
}

.record-button-container {
  display: flex;
  justify-content: center;
  margin-bottom: 20rpx;
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

  &.recording {
    transform: scale(1.1);
    background: linear-gradient(135deg, #FF4757, #FF3838);
  }

  .record-icon {
    font-size: 48rpx;
    margin-bottom: 10rpx;
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
  margin-bottom: 20rpx;

  .recording-time {
    font-size: 32rpx;
    font-weight: bold;
    color: #FF4757;
    margin-bottom: 15rpx;
    display: block;
  }
}

.recording-wave {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 4rpx;

  .wave-bar {
    width: 6rpx;
    height: 20rpx;
    background: #FF4757;
    border-radius: 3rpx;
    animation: wave 1s infinite ease-in-out;

    @for $i from 1 through 20 {
      &:nth-child(#{$i}) {
        animation-delay: #{($i - 1) * 0.05}s;
      }
    }
  }
}

@keyframes wave {
  0%, 100% {
    height: 20rpx;
  }
  50% {
    height: 40rpx;
  }
}

.action-buttons {
  display: flex;
  justify-content: space-around;
  gap: 30rpx;
}

.action-btn {
  flex: 1;
  height: 80rpx;
  background: #f8f9fa;
  border: 1rpx solid #dee2e6;
  border-radius: 15rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10rpx;

  .btn-icon {
    font-size: 28rpx;
  }

  .btn-text {
    font-size: 26rpx;
    color: #666;
  }
}
</style>