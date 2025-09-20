<template>
  <view class="chat-bubble" :class="bubbleClasses">
    <!-- 头像区域 -->
    <view v-if="showAvatar" class="avatar-section" :class="{ 'avatar-right': isUser }">
      <image
        v-if="avatar"
        :src="avatar"
        mode="aspectFill"
        class="avatar-img"
      />
      <view v-else class="avatar-placeholder">
        <text class="avatar-text">{{ avatarText }}</text>
      </view>
    </view>

    <!-- 消息内容区域 -->
    <view class="message-section">
      <!-- 发送者信息 -->
      <view v-if="showSender && senderName" class="sender-info">
        <text class="sender-name">{{ senderName }}</text>
        <text v-if="showTime" class="message-time">{{ formattedTime }}</text>
      </view>

      <!-- 消息气泡 -->
      <view class="bubble-container">
        <!-- 消息状态指示器 -->
        <view v-if="showStatus && isUser" class="message-status" :class="statusClass">
          <text class="status-icon">{{ statusIcon }}</text>
        </view>

        <!-- 气泡内容 -->
        <view class="bubble-content" @longpress="onLongPress">
          <!-- 文本消息 -->
          <view v-if="messageType === 'text'" class="text-content">
            <text class="message-text" :class="{ 'selectable': selectable }">{{ content }}</text>
          </view>

          <!-- 语音消息 -->
          <view v-else-if="messageType === 'audio'" class="audio-content">
            <view class="audio-info">
              <text class="audio-icon">🎤</text>
              <text class="audio-duration">{{ audioDuration || '0:00' }}</text>
            </view>
            <AudioPlayer
              v-if="audioUrl"
              :audioUrl="audioUrl"
              :compact="true"
              :autoPlay="false"
              @play="onAudioPlay"
              @pause="onAudioPause"
              @ended="onAudioEnded"
            />
          </view>

          <!-- 图片消息 -->
          <view v-else-if="messageType === 'image'" class="image-content">
            <image
              :src="imageUrl"
              mode="aspectFill"
              class="message-image"
              @tap="previewImage"
              @error="onImageError"
            />
          </view>

          <!-- 系统消息 -->
          <view v-else-if="messageType === 'system'" class="system-content">
            <text class="system-text">{{ content }}</text>
          </view>

          <!-- 加载中消息 -->
          <view v-else-if="messageType === 'loading'" class="loading-content">
            <view class="loading-dots">
              <text class="dot"></text>
              <text class="dot"></text>
              <text class="dot"></text>
            </view>
            <text class="loading-text">{{ content || 'AI正在思考中...' }}</text>
          </view>

          <!-- 错误消息 -->
          <view v-else-if="messageType === 'error'" class="error-content">
            <text class="error-icon">⚠️</text>
            <text class="error-text">{{ content }}</text>
            <button v-if="retryable" class="retry-btn" @tap="onRetry">
              重试
            </button>
          </view>
        </view>

        <!-- 消息时间（在气泡外显示） -->
        <view v-if="!showSender && showTime" class="bubble-time">
          <text class="time-text">{{ formattedTime }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import AudioPlayer from './AudioPlayer.vue'

interface Props {
  // 消息基本信息
  content: string
  messageType?: 'text' | 'audio' | 'image' | 'system' | 'loading' | 'error'
  isUser?: boolean
  timestamp?: number

  // 发送者信息
  senderName?: string
  avatar?: string
  showSender?: boolean
  showAvatar?: boolean

  // 消息状态
  status?: 'sending' | 'sent' | 'delivered' | 'read' | 'failed'
  showStatus?: boolean
  retryable?: boolean

  // 媒体内容
  audioUrl?: string
  audioDuration?: string
  imageUrl?: string

  // 显示选项
  showTime?: boolean
  selectable?: boolean

  // 样式选项
  maxWidth?: string
  theme?: 'default' | 'compact' | 'minimal'
}

interface Emits {
  (e: 'longpress', message: any): void
  (e: 'retry'): void
  (e: 'audioPlay'): void
  (e: 'audioPause'): void
  (e: 'audioEnded'): void
  (e: 'imagePreview', url: string): void
}

const props = withDefaults(defineProps<Props>(), {
  messageType: 'text',
  isUser: false,
  showSender: false,
  showAvatar: true,
  showStatus: true,
  retryable: false,
  showTime: true,
  selectable: false,
  maxWidth: '70%',
  theme: 'default'
})

const emit = defineEmits<Emits>()

const bubbleClasses = computed(() => {
  return {
    'user-message': props.isUser,
    'ai-message': !props.isUser && props.messageType !== 'system',
    'system-message': props.messageType === 'system',
    [`theme-${props.theme}`]: true,
    [`type-${props.messageType}`]: true
  }
})

const avatarText = computed(() => {
  if (props.isUser) return '我'
  if (props.senderName) return props.senderName.charAt(0)
  return 'AI'
})

const formattedTime = computed(() => {
  if (!props.timestamp) return ''

  const date = new Date(props.timestamp)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMinutes = Math.floor(diffMs / (60 * 1000))

  if (diffMinutes < 1) return '刚刚'
  if (diffMinutes < 60) return `${diffMinutes}分钟前`

  const diffHours = Math.floor(diffMinutes / 60)
  if (diffHours < 24) return `${diffHours}小时前`

  // 显示具体时间
  const hours = date.getHours().toString().padStart(2, '0')
  const minutes = date.getMinutes().toString().padStart(2, '0')

  if (diffMs < 7 * 24 * 60 * 60 * 1000) {
    // 一周内显示星期
    const days = ['日', '一', '二', '三', '四', '五', '六']
    return `周${days[date.getDay()]} ${hours}:${minutes}`
  }

  // 显示日期
  const month = (date.getMonth() + 1).toString().padStart(2, '0')
  const day = date.getDate().toString().padStart(2, '0')
  return `${month}-${day} ${hours}:${minutes}`
})

const statusClass = computed(() => {
  return `status-${props.status}`
})

const statusIcon = computed(() => {
  const iconMap = {
    sending: '⏳',
    sent: '✓',
    delivered: '✓✓',
    read: '✓✓',
    failed: '❌'
  }
  return iconMap[props.status || 'sent'] || '✓'
})

const onLongPress = () => {
  const messageData = {
    content: props.content,
    messageType: props.messageType,
    isUser: props.isUser,
    timestamp: props.timestamp,
    status: props.status
  }
  emit('longpress', messageData)
}

const onRetry = () => {
  emit('retry')
}

const onAudioPlay = () => {
  emit('audioPlay')
}

const onAudioPause = () => {
  emit('audioPause')
}

const onAudioEnded = () => {
  emit('audioEnded')
}

const previewImage = () => {
  if (props.imageUrl) {
    emit('imagePreview', props.imageUrl)

    // 使用uni-app的图片预览
    uni.previewImage({
      urls: [props.imageUrl],
      current: props.imageUrl
    })
  }
}

const onImageError = () => {
  uni.showToast({
    title: '图片加载失败',
    icon: 'none'
  })
}
</script>

<style lang="scss" scoped>
.chat-bubble {
  display: flex;
  margin-bottom: 30rpx;
  align-items: flex-start;

  &.user-message {
    flex-direction: row-reverse;

    .message-section {
      align-items: flex-end;
    }

    .bubble-content {
      background: linear-gradient(135deg, #007AFF, #5856D6);
      color: white;
      margin-right: 20rpx;
    }
  }

  &.ai-message {
    .bubble-content {
      background: white;
      color: #333;
      margin-left: 20rpx;
      box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.1);
    }
  }

  &.system-message {
    justify-content: center;

    .bubble-content {
      background: #f0f0f0;
      color: #666;
      border-radius: 20rpx;
      padding: 15rpx 25rpx;
      font-size: 24rpx;
    }
  }
}

.avatar-section {
  width: 80rpx;
  height: 80rpx;
  border-radius: 40rpx;
  overflow: hidden;
  flex-shrink: 0;

  &.avatar-right {
    margin-left: 20rpx;
  }

  .avatar-img {
    width: 100%;
    height: 100%;
  }

  .avatar-placeholder {
    width: 100%;
    height: 100%;
    background: #ccc;
    display: flex;
    align-items: center;
    justify-content: center;

    .avatar-text {
      color: white;
      font-size: 28rpx;
      font-weight: bold;
    }
  }
}

.message-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  max-width: v-bind(maxWidth);
  min-width: 0;
}

.sender-info {
  display: flex;
  align-items: center;
  gap: 15rpx;
  margin-bottom: 10rpx;

  .sender-name {
    font-size: 24rpx;
    color: #666;
    font-weight: bold;
  }

  .message-time {
    font-size: 22rpx;
    color: #999;
  }
}

.bubble-container {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: flex-start;

  .user-message & {
    align-items: flex-end;
  }
}

.message-status {
  position: absolute;
  right: -35rpx;
  bottom: 10rpx;
  z-index: 1;

  .status-icon {
    font-size: 20rpx;

    &.status-sending {
      color: #999;
    }

    &.status-sent {
      color: #ccc;
    }

    &.status-delivered, &.status-read {
      color: #007AFF;
    }

    &.status-failed {
      color: #ff4757;
    }
  }
}

.bubble-content {
  border-radius: 20rpx;
  padding: 20rpx 25rpx;
  max-width: 100%;
  word-wrap: break-word;
  position: relative;

  .user-message & {
    border-bottom-right-radius: 8rpx;
  }

  .ai-message & {
    border-bottom-left-radius: 8rpx;
  }
}

.text-content {
  .message-text {
    font-size: 30rpx;
    line-height: 1.4;

    &.selectable {
      user-select: text;
    }
  }
}

.audio-content {
  .audio-info {
    display: flex;
    align-items: center;
    gap: 10rpx;
    margin-bottom: 15rpx;

    .audio-icon {
      font-size: 24rpx;
    }

    .audio-duration {
      font-size: 22rpx;
      opacity: 0.8;
    }
  }
}

.image-content {
  .message-image {
    max-width: 400rpx;
    max-height: 400rpx;
    border-radius: 10rpx;
  }
}

.system-content {
  .system-text {
    font-size: 24rpx;
    text-align: center;
  }
}

.loading-content {
  display: flex;
  align-items: center;
  gap: 15rpx;

  .loading-dots {
    display: flex;
    gap: 8rpx;

    .dot {
      width: 8rpx;
      height: 8rpx;
      border-radius: 50%;
      background-color: currentColor;
      opacity: 0.6;
      animation: loading-pulse 1.4s infinite ease-in-out;

      &:nth-child(1) { animation-delay: -0.32s; }
      &:nth-child(2) { animation-delay: -0.16s; }
      &:nth-child(3) { animation-delay: 0s; }
    }
  }

  .loading-text {
    font-size: 26rpx;
    opacity: 0.8;
  }
}

.error-content {
  display: flex;
  align-items: center;
  gap: 15rpx;

  .error-icon {
    font-size: 24rpx;
    color: #ff4757;
  }

  .error-text {
    font-size: 26rpx;
    color: #ff4757;
    flex: 1;
  }

  .retry-btn {
    padding: 8rpx 16rpx;
    background: #ff4757;
    color: white;
    border: none;
    border-radius: 8rpx;
    font-size: 22rpx;
  }
}

.bubble-time {
  margin-top: 8rpx;
  text-align: center;

  .user-message & {
    text-align: right;
  }

  .time-text {
    font-size: 20rpx;
    color: #999;
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

// 主题样式
.theme-compact {
  margin-bottom: 20rpx;

  .bubble-content {
    padding: 15rpx 20rpx;
    border-radius: 15rpx;
  }

  .text-content .message-text {
    font-size: 28rpx;
  }
}

.theme-minimal {
  margin-bottom: 15rpx;

  .bubble-content {
    padding: 12rpx 18rpx;
    border-radius: 12rpx;
    box-shadow: none;
    border: 1rpx solid #eee;
  }

  .text-content .message-text {
    font-size: 26rpx;
  }
}
</style>