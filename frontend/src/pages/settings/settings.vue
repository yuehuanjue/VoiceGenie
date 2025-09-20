<template>
  <view class="settings-container">
    <!-- 用户信息区域 -->
    <view class="user-section">
      <view class="user-avatar" @tap="changeAvatar">
        <image
          v-if="userInfo.avatar"
          :src="userInfo.avatar"
          mode="aspectFill"
          class="avatar-img"
        />
        <view v-else class="avatar-placeholder">
          <text class="avatar-text">{{ userInfo.nickname?.charAt(0) || '用' }}</text>
        </view>
      </view>
      <view class="user-info">
        <view class="user-name" @tap="editProfile">{{ userInfo.nickname || '点击设置昵称' }}</view>
        <view class="user-id">ID: {{ userInfo.id || 'guest' }}</view>
      </view>
      <view class="edit-btn" @tap="editProfile">
        <text class="edit-icon">✏️</text>
      </view>
    </view>

    <!-- 设置列表 -->
    <scroll-view class="settings-list" scroll-y="true">
      <!-- 通用设置 -->
      <view class="setting-group">
        <view class="group-title">通用设置</view>

        <view class="setting-item" @tap="toggleAutoRecord">
          <view class="item-left">
            <text class="item-icon">🎤</text>
            <view class="item-content">
              <text class="item-title">自动开始录音</text>
              <text class="item-desc">进入对话页面时自动开始录音</text>
            </view>
          </view>
          <switch :checked="settings.autoRecord" @change="onAutoRecordChange" />
        </view>

        <view class="setting-item" @tap="selectLanguage">
          <view class="item-left">
            <text class="item-icon">🌐</text>
            <view class="item-content">
              <text class="item-title">语言设置</text>
              <text class="item-desc">{{ currentLanguage }}</text>
            </view>
          </view>
          <text class="item-arrow">›</text>
        </view>

        <view class="setting-item" @tap="selectTheme">
          <view class="item-left">
            <text class="item-icon">🎨</text>
            <view class="item-content">
              <text class="item-title">主题设置</text>
              <text class="item-desc">{{ currentTheme }}</text>
            </view>
          </view>
          <text class="item-arrow">›</text>
        </view>
      </view>

      <!-- 音频设置 -->
      <view class="setting-group">
        <view class="group-title">音频设置</view>

        <view class="setting-item">
          <view class="item-left">
            <text class="item-icon">🔊</text>
            <view class="item-content">
              <text class="item-title">播放音量</text>
              <text class="item-desc">{{ settings.playbackVolume }}%</text>
            </view>
          </view>
          <view class="volume-control">
            <slider
              :value="settings.playbackVolume"
              min="0"
              max="100"
              step="5"
              block-size="18"
              @change="onVolumeChange"
            />
          </view>
        </view>

        <view class="setting-item">
          <view class="item-left">
            <text class="item-icon">🎙️</text>
            <view class="item-content">
              <text class="item-title">录音质量</text>
              <text class="item-desc">{{ getRecordQualityText(settings.recordQuality) }}</text>
            </view>
          </view>
          <picker
            :range="recordQualityOptions"
            :range-key="'label'"
            :value="recordQualityIndex"
            @change="onRecordQualityChange"
          >
            <text class="item-arrow">›</text>
          </picker>
        </view>

        <view class="setting-item" @tap="toggleNoiseReduction">
          <view class="item-left">
            <text class="item-icon">🔇</text>
            <view class="item-content">
              <text class="item-title">降噪处理</text>
              <text class="item-desc">开启智能降噪功能</text>
            </view>
          </view>
          <switch :checked="settings.noiseReduction" @change="onNoiseReductionChange" />
        </view>
      </view>

      <!-- AI 设置 -->
      <view class="setting-group">
        <view class="group-title">AI 设置</view>

        <view class="setting-item" @tap="selectVoice">
          <view class="item-left">
            <text class="item-icon">🗣️</text>
            <view class="item-content">
              <text class="item-title">AI 语音</text>
              <text class="item-desc">{{ currentVoice }}</text>
            </view>
          </view>
          <text class="item-arrow">›</text>
        </view>

        <view class="setting-item">
          <view class="item-left">
            <text class="item-icon">⚡</text>
            <view class="item-content">
              <text class="item-title">响应速度</text>
              <text class="item-desc">{{ getResponseSpeedText(settings.responseSpeed) }}</text>
            </view>
          </view>
          <view class="speed-control">
            <slider
              :value="settings.responseSpeed"
              min="1"
              max="3"
              step="1"
              block-size="18"
              @change="onResponseSpeedChange"
            />
          </view>
        </view>

        <view class="setting-item" @tap="toggleAutoSpeak">
          <view class="item-left">
            <text class="item-icon">📢</text>
            <view class="item-content">
              <text class="item-title">自动播放回复</text>
              <text class="item-desc">AI回复后自动播放语音</text>
            </view>
          </view>
          <switch :checked="settings.autoSpeak" @change="onAutoSpeakChange" />
        </view>
      </view>

      <!-- 隐私设置 -->
      <view class="setting-group">
        <view class="group-title">隐私设置</view>

        <view class="setting-item" @tap="toggleSaveHistory">
          <view class="item-left">
            <text class="item-icon">💾</text>
            <view class="item-content">
              <text class="item-title">保存对话记录</text>
              <text class="item-desc">在设备上保存对话历史</text>
            </view>
          </view>
          <switch :checked="settings.saveHistory" @change="onSaveHistoryChange" />
        </view>

        <view class="setting-item" @tap="clearHistory">
          <view class="item-left">
            <text class="item-icon">🗑️</text>
            <view class="item-content">
              <text class="item-title">清空历史记录</text>
              <text class="item-desc">删除所有本地对话记录</text>
            </view>
          </view>
          <text class="item-arrow">›</text>
        </view>
      </view>

      <!-- 其他设置 -->
      <view class="setting-group">
        <view class="group-title">其他</view>

        <view class="setting-item" @tap="checkUpdate">
          <view class="item-left">
            <text class="item-icon">🔄</text>
            <view class="item-content">
              <text class="item-title">检查更新</text>
              <text class="item-desc">当前版本 v{{ appVersion }}</text>
            </view>
          </view>
          <text class="item-arrow">›</text>
        </view>

        <view class="setting-item" @tap="showHelp">
          <view class="item-left">
            <text class="item-icon">❓</text>
            <view class="item-content">
              <text class="item-title">帮助与反馈</text>
              <text class="item-desc">使用帮助和问题反馈</text>
            </view>
          </view>
          <text class="item-arrow">›</text>
        </view>

        <view class="setting-item" @tap="showAbout">
          <view class="item-left">
            <text class="item-icon">ℹ️</text>
            <view class="item-content">
              <text class="item-title">关于应用</text>
              <text class="item-desc">了解 VoiceGenie</text>
            </view>
          </view>
          <text class="item-arrow">›</text>
        </view>
      </view>

      <!-- 账户操作 -->
      <view class="setting-group">
        <view class="action-buttons">
          <button class="action-btn logout" @tap="logout">
            <text class="btn-text">退出登录</text>
          </button>
        </view>
      </view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

interface UserInfo {
  id: string
  nickname: string
  avatar: string
}

interface Settings {
  autoRecord: boolean
  language: string
  theme: string
  playbackVolume: number
  recordQuality: number
  noiseReduction: boolean
  voiceType: string
  responseSpeed: number
  autoSpeak: boolean
  saveHistory: boolean
}

const userInfo = ref<UserInfo>({
  id: 'guest_001',
  nickname: '语音用户',
  avatar: ''
})

const settings = ref<Settings>({
  autoRecord: false,
  language: 'zh-CN',
  theme: 'auto',
  playbackVolume: 80,
  recordQuality: 2,
  noiseReduction: true,
  voiceType: 'female_01',
  responseSpeed: 2,
  autoSpeak: true,
  saveHistory: true
})

const appVersion = ref<string>('0.1.0')

const recordQualityOptions = [
  { value: 1, label: '标准质量（省流量）' },
  { value: 2, label: '高质量（推荐）' },
  { value: 3, label: '超高质量（耗流量）' }
]

const recordQualityIndex = computed(() => {
  return recordQualityOptions.findIndex(option => option.value === settings.value.recordQuality)
})

const currentLanguage = computed(() => {
  const languages: Record<string, string> = {
    'zh-CN': '简体中文',
    'zh-TW': '繁体中文',
    'en-US': 'English',
    'ja-JP': '日本語'
  }
  return languages[settings.value.language] || '简体中文'
})

const currentTheme = computed(() => {
  const themes: Record<string, string> = {
    'light': '浅色模式',
    'dark': '深色模式',
    'auto': '跟随系统'
  }
  return themes[settings.value.theme] || '跟随系统'
})

const currentVoice = computed(() => {
  const voices: Record<string, string> = {
    'female_01': '温柔女声',
    'female_02': '甜美女声',
    'male_01': '磁性男声',
    'male_02': '沉稳男声'
  }
  return voices[settings.value.voiceType] || '温柔女声'
})

onMounted(() => {
  loadSettings()
  loadUserInfo()
})

const loadSettings = () => {
  try {
    const savedSettings = uni.getStorageSync('app_settings')
    if (savedSettings) {
      settings.value = { ...settings.value, ...savedSettings }
    }
  } catch (error) {
    console.error('加载设置失败:', error)
  }
}

const saveSettings = () => {
  try {
    uni.setStorageSync('app_settings', settings.value)
  } catch (error) {
    console.error('保存设置失败:', error)
  }
}

const loadUserInfo = () => {
  try {
    const savedUserInfo = uni.getStorageSync('user_info')
    if (savedUserInfo) {
      userInfo.value = { ...userInfo.value, ...savedUserInfo }
    }
  } catch (error) {
    console.error('加载用户信息失败:', error)
  }
}

const changeAvatar = () => {
  uni.chooseImage({
    count: 1,
    sizeType: ['compressed'],
    sourceType: ['album', 'camera'],
    success: (res) => {
      userInfo.value.avatar = res.tempFilePaths[0]
      saveUserInfo()
    }
  })
}

const editProfile = () => {
  uni.navigateTo({
    url: '/pages/profile/profile'
  })
}

const saveUserInfo = () => {
  try {
    uni.setStorageSync('user_info', userInfo.value)
  } catch (error) {
    console.error('保存用户信息失败:', error)
  }
}

// 设置项处理函数
const onAutoRecordChange = (e: any) => {
  settings.value.autoRecord = e.detail.value
  saveSettings()
}

const onVolumeChange = (e: any) => {
  settings.value.playbackVolume = e.detail.value
  saveSettings()
}

const onRecordQualityChange = (e: any) => {
  settings.value.recordQuality = recordQualityOptions[e.detail.value].value
  saveSettings()
}

const onNoiseReductionChange = (e: any) => {
  settings.value.noiseReduction = e.detail.value
  saveSettings()
}

const onResponseSpeedChange = (e: any) => {
  settings.value.responseSpeed = e.detail.value
  saveSettings()
}

const onAutoSpeakChange = (e: any) => {
  settings.value.autoSpeak = e.detail.value
  saveSettings()
}

const onSaveHistoryChange = (e: any) => {
  settings.value.saveHistory = e.detail.value
  saveSettings()
}

const selectLanguage = () => {
  const languages = ['简体中文', 'English', '日本語']
  uni.showActionSheet({
    itemList: languages,
    success: (res) => {
      const langMap = ['zh-CN', 'en-US', 'ja-JP']
      settings.value.language = langMap[res.tapIndex]
      saveSettings()
    }
  })
}

const selectTheme = () => {
  const themes = ['跟随系统', '浅色模式', '深色模式']
  uni.showActionSheet({
    itemList: themes,
    success: (res) => {
      const themeMap = ['auto', 'light', 'dark']
      settings.value.theme = themeMap[res.tapIndex]
      saveSettings()
    }
  })
}

const selectVoice = () => {
  const voices = ['温柔女声', '甜美女声', '磁性男声', '沉稳男声']
  uni.showActionSheet({
    itemList: voices,
    success: (res) => {
      const voiceMap = ['female_01', 'female_02', 'male_01', 'male_02']
      settings.value.voiceType = voiceMap[res.tapIndex]
      saveSettings()
    }
  })
}

const getRecordQualityText = (quality: number): string => {
  return recordQualityOptions.find(option => option.value === quality)?.label || '高质量'
}

const getResponseSpeedText = (speed: number): string => {
  const speedTexts: Record<number, string> = {
    1: '慢速（准确优先）',
    2: '标准（推荐）',
    3: '快速（速度优先）'
  }
  return speedTexts[speed] || '标准'
}

const clearHistory = () => {
  uni.showModal({
    title: '确认清空',
    content: '确定要清空所有对话历史记录吗？此操作无法撤销。',
    success: (res) => {
      if (res.confirm) {
        try {
          uni.removeStorageSync('conversation_history')
          uni.showToast({
            title: '清空成功',
            icon: 'success'
          })
        } catch (error) {
          uni.showToast({
            title: '清空失败',
            icon: 'none'
          })
        }
      }
    }
  })
}

const checkUpdate = () => {
  uni.showLoading({ title: '检查中...' })

  setTimeout(() => {
    uni.hideLoading()
    uni.showToast({
      title: '已是最新版本',
      icon: 'success'
    })
  }, 1500)
}

const showHelp = () => {
  uni.navigateTo({
    url: '/pages/help/help'
  })
}

const showAbout = () => {
  uni.navigateTo({
    url: '/pages/about/about'
  })
}

const logout = () => {
  uni.showModal({
    title: '确认退出',
    content: '确定要退出当前账户吗？',
    success: (res) => {
      if (res.confirm) {
        // 清除用户信息
        try {
          uni.removeStorageSync('user_info')
          uni.removeStorageSync('user_token')

          // 跳转到登录页
          uni.redirectTo({
            url: '/pages/login/login'
          })
        } catch (error) {
          uni.showToast({
            title: '退出失败',
            icon: 'none'
          })
        }
      }
    }
  })
}
</script>

<style lang="scss" scoped>
.settings-container {
  height: 100vh;
  background-color: #f5f5f5;
  display: flex;
  flex-direction: column;
}

.user-section {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 60rpx 30rpx 40rpx;
  display: flex;
  align-items: center;
  gap: 30rpx;

  .user-avatar {
    width: 120rpx;
    height: 120rpx;
    border-radius: 60rpx;
    overflow: hidden;
    border: 4rpx solid rgba(255, 255, 255, 0.3);

    .avatar-img {
      width: 100%;
      height: 100%;
    }

    .avatar-placeholder {
      width: 100%;
      height: 100%;
      background: rgba(255, 255, 255, 0.2);
      display: flex;
      align-items: center;
      justify-content: center;

      .avatar-text {
        font-size: 48rpx;
        font-weight: bold;
        color: white;
      }
    }
  }

  .user-info {
    flex: 1;

    .user-name {
      font-size: 36rpx;
      font-weight: bold;
      color: white;
      margin-bottom: 10rpx;
    }

    .user-id {
      font-size: 24rpx;
      color: rgba(255, 255, 255, 0.8);
    }
  }

  .edit-btn {
    width: 60rpx;
    height: 60rpx;
    background: rgba(255, 255, 255, 0.2);
    border-radius: 30rpx;
    display: flex;
    align-items: center;
    justify-content: center;

    .edit-icon {
      font-size: 24rpx;
    }
  }
}

.settings-list {
  flex: 1;
  padding: 30rpx 0;
}

.setting-group {
  margin-bottom: 40rpx;

  .group-title {
    font-size: 26rpx;
    color: #999;
    padding: 0 30rpx 20rpx;
    font-weight: bold;
  }
}

.setting-item {
  background: white;
  padding: 30rpx;
  margin-bottom: 1rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;

  .item-left {
    display: flex;
    align-items: center;
    gap: 20rpx;
    flex: 1;

    .item-icon {
      font-size: 32rpx;
      width: 40rpx;
      text-align: center;
    }

    .item-content {
      flex: 1;

      .item-title {
        font-size: 32rpx;
        color: #333;
        margin-bottom: 5rpx;
        display: block;
      }

      .item-desc {
        font-size: 24rpx;
        color: #999;
        display: block;
      }
    }
  }

  .item-arrow {
    font-size: 36rpx;
    color: #ccc;
    font-weight: 300;
  }

  .volume-control, .speed-control {
    width: 200rpx;
  }
}

.action-buttons {
  padding: 0 30rpx;

  .action-btn {
    width: 100%;
    height: 80rpx;
    border-radius: 15rpx;
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;

    &.logout {
      background: #ff4757;
      color: white;

      .btn-text {
        color: white;
        font-size: 32rpx;
        font-weight: bold;
      }
    }
  }
}
</style>