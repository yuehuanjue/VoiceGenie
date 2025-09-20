<template>
  <view class="login-container">
    <!-- 背景装饰 -->
    <view class="background-decoration">
      <view class="decoration-circle circle-1"></view>
      <view class="decoration-circle circle-2"></view>
      <view class="decoration-circle circle-3"></view>
    </view>

    <!-- 头部区域 -->
    <view class="header-section">
      <view class="logo-area">
        <image src="/static/logo.png" mode="aspectFit" class="logo" />
        <view class="app-name">VoiceGenie</view>
        <view class="app-slogan">您的智能语音助手</view>
      </view>
    </view>

    <!-- 登录表单 -->
    <view class="form-section">
      <view class="form-container">
        <!-- 登录方式切换 -->
        <view class="login-tabs">
          <view
            class="tab-item"
            :class="{ active: loginType === 'phone' }"
            @tap="switchLoginType('phone')"
          >
            手机登录
          </view>
          <view
            class="tab-item"
            :class="{ active: loginType === 'guest' }"
            @tap="switchLoginType('guest')"
          >
            游客体验
          </view>
        </view>

        <!-- 手机登录表单 -->
        <view v-if="loginType === 'phone'" class="phone-form">
          <view class="input-group">
            <view class="input-item">
              <text class="input-icon">📱</text>
              <input
                class="input-field"
                type="number"
                placeholder="请输入手机号"
                v-model="phoneForm.phone"
                maxlength="11"
              />
            </view>

            <view class="input-item">
              <text class="input-icon">🔐</text>
              <input
                class="input-field"
                type="number"
                placeholder="请输入验证码"
                v-model="phoneForm.code"
                maxlength="6"
              />
              <button
                class="code-btn"
                :class="{ disabled: codeCountdown > 0 }"
                @tap="sendCode"
                :disabled="codeCountdown > 0"
              >
                {{ codeCountdown > 0 ? `${codeCountdown}s` : '获取验证码' }}
              </button>
            </view>
          </view>

          <button class="login-btn" @tap="phoneLogin" :disabled="isLogging">
            <text v-if="isLogging">登录中...</text>
            <text v-else>登录</text>
          </button>
        </view>

        <!-- 游客登录 -->
        <view v-if="loginType === 'guest'" class="guest-form">
          <view class="guest-info">
            <view class="guest-icon">👤</view>
            <view class="guest-title">游客模式</view>
            <view class="guest-desc">
              无需注册，立即体验 VoiceGenie 的强大功能
            </view>
            <view class="guest-features">
              <view class="feature-item">
                <text class="feature-icon">✅</text>
                <text class="feature-text">完整的语音对话功能</text>
              </view>
              <view class="feature-item">
                <text class="feature-icon">✅</text>
                <text class="feature-text">本地数据保存</text>
              </view>
              <view class="feature-item">
                <text class="feature-icon">⚠️</text>
                <text class="feature-text">数据不会云端同步</text>
              </view>
            </view>
          </view>

          <button class="guest-btn" @tap="guestLogin" :disabled="isLogging">
            <text v-if="isLogging">进入中...</text>
            <text v-else>立即体验</text>
          </button>
        </view>

        <!-- 第三方登录 -->
        <view class="third-party-section">
          <view class="divider">
            <text class="divider-text">其他登录方式</text>
          </view>

          <view class="third-party-buttons">
            <!-- 微信登录 -->
            <!-- #ifdef MP-WEIXIN -->
            <button class="third-btn wechat" open-type="getUserInfo" @getuserinfo="wechatLogin">
              <text class="third-icon">💬</text>
              <text class="third-text">微信登录</text>
            </button>
            <!-- #endif -->

            <!-- 支付宝登录 -->
            <!-- #ifdef MP-ALIPAY -->
            <button class="third-btn alipay" @tap="alipayLogin">
              <text class="third-icon">💰</text>
              <text class="third-text">支付宝登录</text>
            </button>
            <!-- #endif -->

            <!-- Apple 登录 -->
            <!-- #ifdef APP-PLUS -->
            <button class="third-btn apple" @tap="appleLogin">
              <text class="third-icon">🍎</text>
              <text class="third-text">Apple 登录</text>
            </button>
            <!-- #endif -->
          </view>
        </view>

        <!-- 用户协议 -->
        <view class="agreement-section">
          <view class="agreement-checkbox" @tap="toggleAgreement">
            <text class="checkbox" :class="{ checked: hasAgreed }">
              {{ hasAgreed ? '☑️' : '☐' }}
            </text>
            <text class="agreement-text">
              我已阅读并同意
              <text class="link" @tap.stop="showUserAgreement">《用户协议》</text>
              和
              <text class="link" @tap.stop="showPrivacyPolicy">《隐私政策》</text>
            </text>
          </view>
        </view>
      </view>
    </view>

    <!-- 底部信息 -->
    <view class="footer-section">
      <view class="version-info">v{{ appVersion }}</view>
      <view class="company-info">© 2024 VoiceGenie Team</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const loginType = ref<'phone' | 'guest'>('phone')
const isLogging = ref<boolean>(false)
const hasAgreed = ref<boolean>(false)
const codeCountdown = ref<number>(0)
const appVersion = ref<string>('0.1.0')

const phoneForm = ref({
  phone: '',
  code: ''
})

let countdownTimer: any = null

onMounted(() => {
  checkAutoLogin()
})

const checkAutoLogin = () => {
  try {
    const token = uni.getStorageSync('user_token')
    if (token) {
      // 有token，尝试自动登录
      autoLogin(token)
    }
  } catch (error) {
    console.error('检查自动登录失败:', error)
  }
}

const autoLogin = async (token: string) => {
  try {
    // 这里会调用API验证token
    // const response = await api.verifyToken(token)

    // 模拟验证成功
    setTimeout(() => {
      enterApp()
    }, 1000)
  } catch (error) {
    // token无效，清除本地存储
    uni.removeStorageSync('user_token')
    uni.removeStorageSync('user_info')
  }
}

const switchLoginType = (type: 'phone' | 'guest') => {
  loginType.value = type
}

const sendCode = () => {
  if (!phoneForm.value.phone) {
    uni.showToast({
      title: '请输入手机号',
      icon: 'none'
    })
    return
  }

  if (!/^1[3-9]\d{9}$/.test(phoneForm.value.phone)) {
    uni.showToast({
      title: '手机号格式错误',
      icon: 'none'
    })
    return
  }

  if (!hasAgreed.value) {
    uni.showToast({
      title: '请先同意用户协议',
      icon: 'none'
    })
    return
  }

  // 开始倒计时
  codeCountdown.value = 60
  countdownTimer = setInterval(() => {
    codeCountdown.value--
    if (codeCountdown.value <= 0) {
      clearInterval(countdownTimer)
    }
  }, 1000)

  // 发送验证码
  uni.showToast({
    title: '验证码已发送',
    icon: 'success'
  })

  // 这里会调用API发送验证码
  // api.sendSmsCode(phoneForm.value.phone)
}

const phoneLogin = async () => {
  if (!phoneForm.value.phone || !phoneForm.value.code) {
    uni.showToast({
      title: '请填写完整信息',
      icon: 'none'
    })
    return
  }

  if (!hasAgreed.value) {
    uni.showToast({
      title: '请先同意用户协议',
      icon: 'none'
    })
    return
  }

  isLogging.value = true

  try {
    // 这里会调用API进行手机号登录
    // const response = await api.phoneLogin({
    //   phone: phoneForm.value.phone,
    //   code: phoneForm.value.code
    // })

    // 模拟登录成功
    setTimeout(() => {
      const userInfo = {
        id: 'user_' + Date.now(),
        phone: phoneForm.value.phone,
        nickname: '手机用户',
        avatar: '',
        loginType: 'phone'
      }

      const token = 'token_' + Date.now()

      // 保存用户信息
      uni.setStorageSync('user_info', userInfo)
      uni.setStorageSync('user_token', token)

      uni.showToast({
        title: '登录成功',
        icon: 'success'
      })

      enterApp()
    }, 2000)

  } catch (error) {
    console.error('手机登录失败:', error)
    uni.showToast({
      title: '登录失败，请重试',
      icon: 'none'
    })
  } finally {
    isLogging.value = false
  }
}

const guestLogin = () => {
  if (!hasAgreed.value) {
    uni.showToast({
      title: '请先同意用户协议',
      icon: 'none'
    })
    return
  }

  isLogging.value = true

  // 创建游客用户信息
  const guestInfo = {
    id: 'guest_' + Date.now(),
    nickname: '游客用户',
    avatar: '',
    loginType: 'guest'
  }

  // 保存游客信息
  uni.setStorageSync('user_info', guestInfo)

  setTimeout(() => {
    uni.showToast({
      title: '进入成功',
      icon: 'success'
    })
    enterApp()
  }, 1000)
}

const wechatLogin = (e: any) => {
  if (!hasAgreed.value) {
    uni.showToast({
      title: '请先同意用户协议',
      icon: 'none'
    })
    return
  }

  if (e.detail.userInfo) {
    isLogging.value = true

    // 微信授权成功，进行登录
    const userInfo = {
      id: 'wx_' + Date.now(),
      nickname: e.detail.userInfo.nickName,
      avatar: e.detail.userInfo.avatarUrl,
      loginType: 'wechat'
    }

    uni.setStorageSync('user_info', userInfo)

    setTimeout(() => {
      uni.showToast({
        title: '登录成功',
        icon: 'success'
      })
      enterApp()
    }, 1000)
  } else {
    uni.showToast({
      title: '微信授权失败',
      icon: 'none'
    })
  }
}

const alipayLogin = () => {
  // 支付宝登录逻辑
  uni.showToast({
    title: '支付宝登录开发中',
    icon: 'none'
  })
}

const appleLogin = () => {
  // Apple 登录逻辑
  uni.showToast({
    title: 'Apple 登录开发中',
    icon: 'none'
  })
}

const toggleAgreement = () => {
  hasAgreed.value = !hasAgreed.value
}

const showUserAgreement = () => {
  uni.navigateTo({
    url: '/pages/webview/webview?url=https://voicegenie.app/agreement'
  })
}

const showPrivacyPolicy = () => {
  uni.navigateTo({
    url: '/pages/webview/webview?url=https://voicegenie.app/privacy'
  })
}

const enterApp = () => {
  isLogging.value = false

  // 跳转到首页
  uni.switchTab({
    url: '/pages/index/index'
  })
}
</script>

<style lang="scss" scoped>
.login-container {
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  position: relative;
  overflow: hidden;
}

.background-decoration {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;

  .decoration-circle {
    position: absolute;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.1);

    &.circle-1 {
      width: 200rpx;
      height: 200rpx;
      top: 10%;
      right: -50rpx;
      animation: float 6s ease-in-out infinite;
    }

    &.circle-2 {
      width: 150rpx;
      height: 150rpx;
      top: 30%;
      left: -30rpx;
      animation: float 8s ease-in-out infinite reverse;
    }

    &.circle-3 {
      width: 100rpx;
      height: 100rpx;
      bottom: 20%;
      right: 10%;
      animation: float 7s ease-in-out infinite;
    }
  }
}

@keyframes float {
  0%, 100% {
    transform: translateY(0px);
  }
  50% {
    transform: translateY(-20px);
  }
}

.header-section {
  padding: 120rpx 0 80rpx;
  text-align: center;

  .logo-area {
    .logo {
      width: 160rpx;
      height: 160rpx;
      margin-bottom: 30rpx;
    }

    .app-name {
      font-size: 64rpx;
      font-weight: bold;
      color: white;
      margin-bottom: 15rpx;
    }

    .app-slogan {
      font-size: 28rpx;
      color: rgba(255, 255, 255, 0.8);
    }
  }
}

.form-section {
  flex: 1;
  padding: 0 40rpx;

  .form-container {
    background: white;
    border-radius: 30rpx;
    padding: 50rpx 40rpx;
    box-shadow: 0 20rpx 60rpx rgba(0, 0, 0, 0.1);
  }
}

.login-tabs {
  display: flex;
  background: #f8f9fa;
  border-radius: 15rpx;
  margin-bottom: 40rpx;
  padding: 8rpx;

  .tab-item {
    flex: 1;
    text-align: center;
    padding: 20rpx 0;
    font-size: 28rpx;
    color: #666;
    border-radius: 10rpx;
    transition: all 0.3s ease;

    &.active {
      background: white;
      color: #007AFF;
      font-weight: bold;
      box-shadow: 0 2rpx 8rpx rgba(0, 122, 255, 0.2);
    }
  }
}

.phone-form {
  .input-group {
    margin-bottom: 40rpx;

    .input-item {
      display: flex;
      align-items: center;
      background: #f8f9fa;
      border-radius: 15rpx;
      padding: 25rpx 30rpx;
      margin-bottom: 25rpx;

      .input-icon {
        font-size: 32rpx;
        margin-right: 20rpx;
      }

      .input-field {
        flex: 1;
        font-size: 28rpx;
        color: #333;
        border: none;
        outline: none;
        background: transparent;
      }

      .code-btn {
        padding: 12rpx 24rpx;
        background: #007AFF;
        color: white;
        border-radius: 10rpx;
        font-size: 24rpx;
        border: none;

        &.disabled {
          background: #ccc;
        }
      }
    }
  }

  .login-btn {
    width: 100%;
    height: 80rpx;
    background: linear-gradient(45deg, #007AFF, #5856D6);
    color: white;
    border-radius: 15rpx;
    font-size: 32rpx;
    font-weight: bold;
    border: none;
    box-shadow: 0 8rpx 24rpx rgba(0, 122, 255, 0.3);
  }
}

.guest-form {
  .guest-info {
    text-align: center;
    margin-bottom: 40rpx;

    .guest-icon {
      font-size: 80rpx;
      margin-bottom: 20rpx;
    }

    .guest-title {
      font-size: 36rpx;
      font-weight: bold;
      color: #333;
      margin-bottom: 15rpx;
    }

    .guest-desc {
      font-size: 26rpx;
      color: #666;
      line-height: 1.5;
      margin-bottom: 30rpx;
    }

    .guest-features {
      .feature-item {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 10rpx;
        margin-bottom: 15rpx;

        .feature-icon {
          font-size: 24rpx;
        }

        .feature-text {
          font-size: 24rpx;
          color: #666;
        }
      }
    }
  }

  .guest-btn {
    width: 100%;
    height: 80rpx;
    background: linear-gradient(45deg, #FF6B6B, #FF8E8E);
    color: white;
    border-radius: 15rpx;
    font-size: 32rpx;
    font-weight: bold;
    border: none;
    box-shadow: 0 8rpx 24rpx rgba(255, 107, 107, 0.3);
  }
}

.third-party-section {
  margin-top: 40rpx;

  .divider {
    text-align: center;
    margin-bottom: 30rpx;
    position: relative;

    &::before, &::after {
      content: '';
      position: absolute;
      top: 50%;
      width: 100rpx;
      height: 1rpx;
      background: #ddd;
    }

    &::before {
      left: 0;
    }

    &::after {
      right: 0;
    }

    .divider-text {
      font-size: 24rpx;
      color: #999;
      background: white;
      padding: 0 20rpx;
    }
  }

  .third-party-buttons {
    display: flex;
    gap: 20rpx;

    .third-btn {
      flex: 1;
      height: 80rpx;
      border-radius: 15rpx;
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 10rpx;
      border: 1rpx solid #ddd;
      background: white;

      &.wechat {
        background: #1aad19;
        color: white;
        border-color: #1aad19;
      }

      &.alipay {
        background: #1677ff;
        color: white;
        border-color: #1677ff;
      }

      &.apple {
        background: #000;
        color: white;
        border-color: #000;
      }

      .third-icon {
        font-size: 28rpx;
      }

      .third-text {
        font-size: 24rpx;
      }
    }
  }
}

.agreement-section {
  margin-top: 40rpx;

  .agreement-checkbox {
    display: flex;
    align-items: flex-start;
    gap: 15rpx;

    .checkbox {
      font-size: 28rpx;
      color: #ccc;

      &.checked {
        color: #007AFF;
      }
    }

    .agreement-text {
      font-size: 24rpx;
      color: #666;
      line-height: 1.5;
      flex: 1;

      .link {
        color: #007AFF;
        text-decoration: underline;
      }
    }
  }
}

.footer-section {
  text-align: center;
  padding: 40rpx 0 60rpx;

  .version-info {
    font-size: 22rpx;
    color: rgba(255, 255, 255, 0.6);
    margin-bottom: 10rpx;
  }

  .company-info {
    font-size: 22rpx;
    color: rgba(255, 255, 255, 0.6);
  }
}
</style>