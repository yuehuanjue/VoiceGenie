<template>
  <view class="home-container">
    <!-- 头部区域 -->
    <view class="header">
      <view class="logo">
        <image src="/static/logo.png" mode="aspectFit" class="logo-img" />
      </view>
      <view class="title">VoiceGenie</view>
      <view class="subtitle">您的智能语音助手</view>
    </view>

    <!-- 功能卡片区域 -->
    <view class="feature-cards">
      <view class="card" @tap="startVoiceChat">
        <view class="card-icon">🎤</view>
        <view class="card-title">语音对话</view>
        <view class="card-desc">开始与AI进行语音对话</view>
      </view>

      <view class="card" @tap="viewHistory">
        <view class="card-icon">📝</view>
        <view class="card-title">对话历史</view>
        <view class="card-desc">查看历史对话记录</view>
      </view>

      <view class="card" @tap="openSettings">
        <view class="card-icon">⚙️</view>
        <view class="card-title">设置</view>
        <view class="card-desc">个性化设置和偏好</view>
      </view>
    </view>

    <!-- 快速开始按钮 -->
    <view class="quick-start">
      <button class="start-btn" type="primary" @tap="quickStart">
        <view class="btn-content">
          <text class="btn-icon">🚀</text>
          <text class="btn-text">立即开始对话</text>
        </view>
      </button>
    </view>

    <!-- 状态信息 -->
    <view class="status-info">
      <view class="status-item">
        <text class="status-label">服务状态:</text>
        <text class="status-value" :class="{ 'online': isOnline }">
          {{ isOnline ? '在线' : '离线' }}
        </text>
      </view>
      <view class="status-item">
        <text class="status-label">今日对话:</text>
        <text class="status-value">{{ todayChats }}</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

// 响应式数据
const isOnline = ref<boolean>(true)
const todayChats = ref<number>(0)

// 生命周期
onMounted(() => {
  checkServiceStatus()
  loadTodayStats()
})

// 方法
const startVoiceChat = () => {
  uni.navigateTo({
    url: '/pages/chat/chat'
  })
}

const viewHistory = () => {
  uni.switchTab({
    url: '/pages/history/history'
  })
}

const openSettings = () => {
  uni.switchTab({
    url: '/pages/settings/settings'
  })
}

const quickStart = () => {
  // 检查权限
  checkPermissions().then(() => {
    startVoiceChat()
  }).catch(() => {
    uni.showToast({
      title: '请授权麦克风权限',
      icon: 'none'
    })
  })
}

const checkServiceStatus = async () => {
  try {
    // 这里会调用API检查服务状态
    // const response = await api.checkStatus()
    // isOnline.value = response.online
    isOnline.value = true // 临时设置
  } catch (error) {
    isOnline.value = false
  }
}

const loadTodayStats = async () => {
  try {
    // 这里会调用API获取今日统计
    // const response = await api.getTodayStats()
    // todayChats.value = response.chatCount
    todayChats.value = 0 // 临时设置
  } catch (error) {
    console.error('Failed to load today stats:', error)
  }
}

const checkPermissions = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    // #ifdef MP-WEIXIN
    uni.authorize({
      scope: 'scope.record',
      success: () => resolve(),
      fail: () => reject()
    })
    // #endif

    // #ifdef H5
    navigator.mediaDevices.getUserMedia({ audio: true })
      .then(() => resolve())
      .catch(() => reject())
    // #endif

    // #ifdef APP-PLUS
    resolve() // App端在manifest中已配置权限
    // #endif
  })
}
</script>

<style lang="scss" scoped>
.home-container {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 40rpx 30rpx;
}

.header {
  text-align: center;
  margin-bottom: 80rpx;

  .logo {
    margin-bottom: 20rpx;

    .logo-img {
      width: 120rpx;
      height: 120rpx;
    }
  }

  .title {
    font-size: 48rpx;
    font-weight: bold;
    color: #ffffff;
    margin-bottom: 10rpx;
  }

  .subtitle {
    font-size: 28rpx;
    color: rgba(255, 255, 255, 0.8);
  }
}

.feature-cards {
  display: flex;
  flex-direction: column;
  gap: 30rpx;
  margin-bottom: 60rpx;

  .card {
    background: rgba(255, 255, 255, 0.95);
    border-radius: 20rpx;
    padding: 40rpx 30rpx;
    box-shadow: 0 10rpx 30rpx rgba(0, 0, 0, 0.1);
    transition: transform 0.3s ease;

    &:active {
      transform: scale(0.98);
    }

    .card-icon {
      font-size: 60rpx;
      text-align: center;
      margin-bottom: 20rpx;
    }

    .card-title {
      font-size: 32rpx;
      font-weight: bold;
      color: #333333;
      text-align: center;
      margin-bottom: 10rpx;
    }

    .card-desc {
      font-size: 24rpx;
      color: #666666;
      text-align: center;
      line-height: 1.4;
    }
  }
}

.quick-start {
  margin-bottom: 60rpx;

  .start-btn {
    width: 100%;
    height: 100rpx;
    border-radius: 50rpx;
    background: linear-gradient(45deg, #FF6B6B, #FF8E8E);
    border: none;
    box-shadow: 0 10rpx 30rpx rgba(255, 107, 107, 0.3);

    .btn-content {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 15rpx;

      .btn-icon {
        font-size: 32rpx;
      }

      .btn-text {
        font-size: 32rpx;
        font-weight: bold;
        color: #ffffff;
      }
    }
  }
}

.status-info {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 15rpx;
  padding: 30rpx;
  backdrop-filter: blur(10rpx);

  .status-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20rpx;

    &:last-child {
      margin-bottom: 0;
    }

    .status-label {
      font-size: 28rpx;
      color: rgba(255, 255, 255, 0.8);
    }

    .status-value {
      font-size: 28rpx;
      font-weight: bold;
      color: #ffffff;

      &.online {
        color: #4CAF50;
      }
    }
  }
}
</style>