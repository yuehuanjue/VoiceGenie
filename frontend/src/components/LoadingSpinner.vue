<template>
  <view class="loading-spinner" :class="spinnerClasses" :style="spinnerStyle">
    <!-- 默认圆形加载器 -->
    <view v-if="type === 'circle'" class="circle-spinner">
      <view class="circle-border">
        <view class="circle-core"></view>
      </view>
    </view>

    <!-- 点状加载器 -->
    <view v-else-if="type === 'dots'" class="dots-spinner">
      <view
        v-for="i in 3"
        :key="i"
        class="dot"
        :style="{ animationDelay: `${(i - 1) * 0.16}s` }"
      ></view>
    </view>

    <!-- 脉冲加载器 -->
    <view v-else-if="type === 'pulse'" class="pulse-spinner">
      <view
        v-for="i in 3"
        :key="i"
        class="pulse-ring"
        :style="{ animationDelay: `${(i - 1) * 0.2}s` }"
      ></view>
    </view>

    <!-- 波浪加载器 -->
    <view v-else-if="type === 'wave'" class="wave-spinner">
      <view
        v-for="i in 5"
        :key="i"
        class="wave-bar"
        :style="{ animationDelay: `${(i - 1) * 0.1}s` }"
      ></view>
    </view>

    <!-- 旋转方块 -->
    <view v-else-if="type === 'square'" class="square-spinner">
      <view class="square"></view>
    </view>

    <!-- 弹跳球 -->
    <view v-else-if="type === 'bounce'" class="bounce-spinner">
      <view
        v-for="i in 3"
        :key="i"
        class="bounce-ball"
        :style="{ animationDelay: `${(i - 1) * 0.16}s` }"
      ></view>
    </view>

    <!-- 音频波形（语音专用） -->
    <view v-else-if="type === 'audio'" class="audio-spinner">
      <view
        v-for="i in 4"
        :key="i"
        class="audio-bar"
        :style="{ animationDelay: `${(i - 1) * 0.12}s` }"
      ></view>
    </view>

    <!-- 圆环进度条 -->
    <view v-else-if="type === 'progress'" class="progress-spinner">
      <view class="progress-circle">
        <view
          class="progress-fill"
          :style="{ transform: `rotate(${progress * 3.6}deg)` }"
        ></view>
      </view>
      <text v-if="showProgress" class="progress-text">{{ Math.round(progress) }}%</text>
    </view>

    <!-- 自定义emoji加载器 -->
    <view v-else-if="type === 'emoji'" class="emoji-spinner">
      <text class="emoji-icon">{{ emoji }}</text>
    </view>

    <!-- 加载文本 -->
    <view v-if="text && !hideText" class="loading-text" :style="textStyle">
      {{ text }}
    </view>

    <!-- 提示消息 -->
    <view v-if="tip" class="loading-tip">
      {{ tip }}
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  // 加载器类型
  type?: 'circle' | 'dots' | 'pulse' | 'wave' | 'square' | 'bounce' | 'audio' | 'progress' | 'emoji'

  // 尺寸和颜色
  size?: number | string
  color?: string
  backgroundColor?: string

  // 显示选项
  text?: string
  tip?: string
  hideText?: boolean
  showProgress?: boolean

  // 进度（仅对 progress 类型有效）
  progress?: number

  // 自定义emoji（仅对 emoji 类型有效）
  emoji?: string

  // 速度控制
  speed?: 'slow' | 'normal' | 'fast'

  // 主题
  theme?: 'default' | 'dark' | 'light' | 'primary' | 'success' | 'warning' | 'error'

  // 布局
  overlay?: boolean
  center?: boolean
  inline?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  type: 'circle',
  size: 40,
  color: '#007AFF',
  backgroundColor: 'transparent',
  hideText: false,
  showProgress: true,
  progress: 0,
  emoji: '🔄',
  speed: 'normal',
  theme: 'default',
  overlay: false,
  center: true,
  inline: false
})

const spinnerClasses = computed(() => {
  return {
    [`theme-${props.theme}`]: true,
    [`speed-${props.speed}`]: true,
    'overlay': props.overlay,
    'center': props.center,
    'inline': props.inline
  }
})

const spinnerStyle = computed(() => {
  const sizeValue = typeof props.size === 'number' ? `${props.size}rpx` : props.size

  return {
    '--spinner-size': sizeValue,
    '--spinner-color': props.color,
    '--spinner-bg': props.backgroundColor
  }
})

const textStyle = computed(() => {
  return {
    color: props.color
  }
})

// 主题颜色映射
const themeColors = {
  default: '#007AFF',
  dark: '#333333',
  light: '#ffffff',
  primary: '#007AFF',
  success: '#28a745',
  warning: '#ffc107',
  error: '#dc3545'
}
</script>

<style lang="scss" scoped>
.loading-spinner {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 20rpx;

  &.overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    z-index: 9999;
  }

  &.center {
    min-height: 200rpx;
  }

  &.inline {
    flex-direction: row;
    gap: 15rpx;
    min-height: auto;
  }

  // 速度控制
  &.speed-slow {
    --animation-duration: 2s;
  }

  &.speed-normal {
    --animation-duration: 1s;
  }

  &.speed-fast {
    --animation-duration: 0.5s;
  }
}

// 圆形加载器
.circle-spinner {
  width: var(--spinner-size);
  height: var(--spinner-size);

  .circle-border {
    width: 100%;
    height: 100%;
    border: 4rpx solid var(--spinner-bg, #f3f3f3);
    border-top: 4rpx solid var(--spinner-color);
    border-radius: 50%;
    animation: circle-spin var(--animation-duration, 1s) linear infinite;
  }

  .circle-core {
    width: 60%;
    height: 60%;
    background: var(--spinner-color);
    border-radius: 50%;
    margin: 20% auto;
    opacity: 0.3;
    animation: circle-pulse var(--animation-duration, 1s) ease-in-out infinite alternate;
  }
}

// 点状加载器
.dots-spinner {
  display: flex;
  gap: 8rpx;

  .dot {
    width: 12rpx;
    height: 12rpx;
    background: var(--spinner-color);
    border-radius: 50%;
    animation: dots-bounce var(--animation-duration, 1s) infinite;
  }
}

// 脉冲加载器
.pulse-spinner {
  position: relative;
  width: var(--spinner-size);
  height: var(--spinner-size);

  .pulse-ring {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    border: 4rpx solid var(--spinner-color);
    border-radius: 50%;
    opacity: 1;
    animation: pulse-scale var(--animation-duration, 1s) infinite;
  }
}

// 波浪加载器
.wave-spinner {
  display: flex;
  align-items: flex-end;
  gap: 4rpx;
  height: var(--spinner-size);

  .wave-bar {
    width: 6rpx;
    background: var(--spinner-color);
    border-radius: 3rpx;
    animation: wave-height var(--animation-duration, 1s) infinite;
  }
}

// 方块加载器
.square-spinner {
  width: var(--spinner-size);
  height: var(--spinner-size);

  .square {
    width: 100%;
    height: 100%;
    background: var(--spinner-color);
    border-radius: 4rpx;
    animation: square-rotate var(--animation-duration, 1s) infinite;
  }
}

// 弹跳球加载器
.bounce-spinner {
  display: flex;
  gap: 8rpx;

  .bounce-ball {
    width: 16rpx;
    height: 16rpx;
    background: var(--spinner-color);
    border-radius: 50%;
    animation: bounce-up var(--animation-duration, 1s) infinite;
  }
}

// 音频波形加载器
.audio-spinner {
  display: flex;
  align-items: center;
  gap: 4rpx;
  height: var(--spinner-size);

  .audio-bar {
    width: 6rpx;
    background: var(--spinner-color);
    border-radius: 3rpx;
    animation: audio-wave var(--animation-duration, 1s) infinite;

    &:nth-child(1) { height: 20rpx; }
    &:nth-child(2) { height: 30rpx; }
    &:nth-child(3) { height: 25rpx; }
    &:nth-child(4) { height: 15rpx; }
  }
}

// 进度圆环
.progress-spinner {
  position: relative;
  width: var(--spinner-size);
  height: var(--spinner-size);
  display: flex;
  align-items: center;
  justify-content: center;

  .progress-circle {
    width: 100%;
    height: 100%;
    border: 6rpx solid var(--spinner-bg, #f0f0f0);
    border-radius: 50%;
    position: relative;
    overflow: hidden;

    .progress-fill {
      position: absolute;
      top: -6rpx;
      left: -6rpx;
      width: calc(100% + 12rpx);
      height: calc(100% + 12rpx);
      border: 6rpx solid var(--spinner-color);
      border-radius: 50%;
      border-right-color: transparent;
      border-bottom-color: transparent;
      transform-origin: center;
      transition: transform 0.3s ease;
    }
  }

  .progress-text {
    position: absolute;
    font-size: 24rpx;
    font-weight: bold;
    color: var(--spinner-color);
  }
}

// Emoji加载器
.emoji-spinner {
  .emoji-icon {
    font-size: var(--spinner-size);
    animation: emoji-spin var(--animation-duration, 1s) linear infinite;
  }
}

// 加载文本
.loading-text {
  font-size: 28rpx;
  color: var(--spinner-color);
  text-align: center;
  margin-top: 10rpx;
}

// 提示信息
.loading-tip {
  font-size: 24rpx;
  color: #999;
  text-align: center;
  margin-top: 5rpx;
}

// 主题样式
.theme-dark {
  .loading-text {
    color: #ffffff;
  }

  .loading-tip {
    color: #cccccc;
  }
}

.theme-light {
  .loading-text {
    color: #333333;
  }
}

// 动画定义
@keyframes circle-spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

@keyframes circle-pulse {
  0% { transform: scale(0.8); opacity: 0.3; }
  100% { transform: scale(1); opacity: 0.7; }
}

@keyframes dots-bounce {
  0%, 80%, 100% {
    transform: scale(0);
    opacity: 0.5;
  }
  40% {
    transform: scale(1);
    opacity: 1;
  }
}

@keyframes pulse-scale {
  0% {
    transform: scale(0);
    opacity: 1;
  }
  100% {
    transform: scale(1);
    opacity: 0;
  }
}

@keyframes wave-height {
  0%, 40%, 100% {
    height: 12rpx;
  }
  20% {
    height: var(--spinner-size);
  }
}

@keyframes square-rotate {
  0% {
    transform: perspective(120rpx) rotateX(0deg) rotateY(0deg);
  }
  50% {
    transform: perspective(120rpx) rotateX(-180.1deg) rotateY(0deg);
  }
  100% {
    transform: perspective(120rpx) rotateX(-180deg) rotateY(-179.9deg);
  }
}

@keyframes bounce-up {
  0%, 80%, 100% {
    transform: scale(0);
  }
  40% {
    transform: scale(1);
  }
}

@keyframes audio-wave {
  0%, 100% {
    transform: scaleY(0.4);
  }
  50% {
    transform: scaleY(1);
  }
}

@keyframes emoji-spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

// 响应式调整
@media (max-width: 750rpx) {
  .loading-spinner {
    &.overlay {
      padding: 40rpx;
    }
  }

  .loading-text {
    font-size: 26rpx;
  }

  .loading-tip {
    font-size: 22rpx;
  }
}
</style>