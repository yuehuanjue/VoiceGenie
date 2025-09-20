import { http } from './request'
import {
  ApiResponse,
  PaginatedResponse,
  UserInfo,
  Message,
  Conversation,
  ASRResult,
  TTSRequest,
  TTSResult,
  ChatRequest,
  ChatResponse,
  UploadResponse,
  UploadProgressCallback
} from './types'

// 认证相关API
export const authApi = {
  // 手机号验证码登录
  phoneLogin(data: { phone: string; code: string }): Promise<ApiResponse<{ token: string; userInfo: UserInfo }>> {
    return http.post('/auth/phone/login', data)
  },

  // 发送验证码
  sendSmsCode(phone: string): Promise<ApiResponse<{ expireTime: number }>> {
    return http.post('/auth/sms/send', { phone })
  },

  // 微信登录
  wechatLogin(data: { code: string; userInfo: any }): Promise<ApiResponse<{ token: string; userInfo: UserInfo }>> {
    return http.post('/auth/wechat/login', data)
  },

  // 游客登录
  guestLogin(): Promise<ApiResponse<{ token: string; userInfo: UserInfo }>> {
    return http.post('/auth/guest/login')
  },

  // 验证token
  verifyToken(): Promise<ApiResponse<UserInfo>> {
    return http.get('/auth/verify')
  },

  // 刷新token
  refreshToken(refreshToken: string): Promise<ApiResponse<{ token: string }>> {
    return http.post('/auth/refresh', { refreshToken })
  },

  // 登出
  logout(): Promise<ApiResponse<void>> {
    return http.post('/auth/logout')
  }
}

// 用户相关API
export const userApi = {
  // 获取用户信息
  getUserInfo(): Promise<ApiResponse<UserInfo>> {
    return http.get('/user/info')
  },

  // 更新用户信息
  updateUserInfo(data: Partial<UserInfo>): Promise<ApiResponse<UserInfo>> {
    return http.put('/user/info', data)
  },

  // 上传头像
  uploadAvatar(filePath: string, onProgress?: UploadProgressCallback): Promise<ApiResponse<UploadResponse>> {
    return http.upload('/user/avatar', filePath, {
      name: 'avatar',
      onProgress
    })
  },

  // 绑定手机号
  bindPhone(data: { phone: string; code: string }): Promise<ApiResponse<void>> {
    return http.post('/user/bind/phone', data)
  },

  // 修改密码
  changePassword(data: { oldPassword: string; newPassword: string }): Promise<ApiResponse<void>> {
    return http.post('/user/password/change', data)
  }
}

// 对话相关API
export const conversationApi = {
  // 获取对话列表
  getConversations(params: {
    page?: number
    pageSize?: number
    keyword?: string
  } = {}): Promise<ApiResponse<PaginatedResponse<Conversation>>> {
    return http.get('/conversations', params)
  },

  // 获取对话详情
  getConversation(id: string): Promise<ApiResponse<Conversation>> {
    return http.get(`/conversations/${id}`)
  },

  // 创建对话
  createConversation(data: { title?: string }): Promise<ApiResponse<Conversation>> {
    return http.post('/conversations', data)
  },

  // 更新对话
  updateConversation(id: string, data: { title?: string }): Promise<ApiResponse<Conversation>> {
    return http.put(`/conversations/${id}`, data)
  },

  // 删除对话
  deleteConversation(id: string): Promise<ApiResponse<void>> {
    return http.delete(`/conversations/${id}`)
  },

  // 清空所有对话
  clearConversations(): Promise<ApiResponse<void>> {
    return http.delete('/conversations/all')
  }
}

// 消息相关API
export const messageApi = {
  // 获取消息列表
  getMessages(conversationId: string, params: {
    page?: number
    pageSize?: number
    lastMessageId?: string
  } = {}): Promise<ApiResponse<PaginatedResponse<Message>>> {
    return http.get(`/conversations/${conversationId}/messages`, params)
  },

  // 发送文本消息
  sendTextMessage(data: {
    conversationId: string
    content: string
  }): Promise<ApiResponse<Message>> {
    return http.post('/messages/text', data)
  },

  // 发送语音消息
  sendVoiceMessage(data: {
    conversationId: string
    audioUrl: string
    duration: number
  }): Promise<ApiResponse<Message>> {
    return http.post('/messages/voice', data)
  },

  // 删除消息
  deleteMessage(messageId: string): Promise<ApiResponse<void>> {
    return http.delete(`/messages/${messageId}`)
  }
}

// 语音处理相关API
export const voiceApi = {
  // 上传音频文件
  uploadAudio(filePath: string, onProgress?: UploadProgressCallback): Promise<ApiResponse<UploadResponse>> {
    return http.upload('/voice/upload', filePath, {
      name: 'audio',
      onProgress
    })
  },

  // 语音转文字 (ASR)
  speechToText(audioUrl: string, options?: {
    language?: string
    enablePunctuation?: boolean
    enableWordTimeStamp?: boolean
  }): Promise<ApiResponse<ASRResult>> {
    return http.post('/voice/asr', {
      audioUrl,
      ...options
    })
  },

  // 文字转语音 (TTS)
  textToSpeech(request: TTSRequest): Promise<ApiResponse<TTSResult>> {
    return http.post('/voice/tts', request)
  },

  // 获取支持的语音列表
  getVoiceList(): Promise<ApiResponse<Array<{
    id: string
    name: string
    language: string
    gender: 'male' | 'female'
    description: string
  }>>> {
    return http.get('/voice/voices')
  }
}

// AI聊天相关API
export const chatApi = {
  // 发送聊天消息
  sendMessage(request: ChatRequest): Promise<ApiResponse<ChatResponse>> {
    return http.post('/chat/send', request)
  },

  // 流式聊天（SSE）
  sendMessageStream(request: ChatRequest, onMessage: (chunk: string) => void): Promise<void> {
    return new Promise((resolve, reject) => {
      // #ifdef H5
      const eventSource = new EventSource(`${http.config.baseURL}/chat/stream`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json'
        }
      })

      eventSource.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          onMessage(data.content)
        } catch (error) {
          console.error('解析SSE消息失败:', error)
        }
      }

      eventSource.onerror = (error) => {
        eventSource.close()
        reject(error)
      }

      eventSource.addEventListener('end', () => {
        eventSource.close()
        resolve()
      })

      // 发送初始消息
      fetch(`${http.config.baseURL}/chat/stream`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(request)
      })
      // #endif

      // #ifdef MP-WEIXIN || APP-PLUS
      // 小程序和App使用WebSocket实现流式聊天
      const socketTask = uni.connectSocket({
        url: 'wss://api.voicegenie.app/chat/ws',
        header: {
          'Authorization': `Bearer ${uni.getStorageSync('user_token')}`
        }
      })

      socketTask.onOpen(() => {
        socketTask.send({
          data: JSON.stringify(request)
        })
      })

      socketTask.onMessage((res) => {
        try {
          const data = JSON.parse(res.data as string)
          if (data.type === 'message') {
            onMessage(data.content)
          } else if (data.type === 'end') {
            socketTask.close()
            resolve()
          }
        } catch (error) {
          console.error('解析WebSocket消息失败:', error)
        }
      })

      socketTask.onError((error) => {
        socketTask.close()
        reject(error)
      })
      // #endif
    })
  },

  // 获取聊天建议
  getSuggestions(conversationId: string): Promise<ApiResponse<string[]>> {
    return http.get(`/chat/suggestions/${conversationId}`)
  },

  // 清空聊天上下文
  clearContext(conversationId: string): Promise<ApiResponse<void>> {
    return http.delete(`/chat/context/${conversationId}`)
  }
}

// 设置相关API
export const settingsApi = {
  // 获取用户设置
  getSettings(): Promise<ApiResponse<Record<string, any>>> {
    return http.get('/settings')
  },

  // 更新设置
  updateSettings(settings: Record<string, any>): Promise<ApiResponse<void>> {
    return http.put('/settings', settings)
  },

  // 重置设置
  resetSettings(): Promise<ApiResponse<void>> {
    return http.delete('/settings')
  }
}

// 系统相关API
export const systemApi = {
  // 检查服务状态
  checkStatus(): Promise<ApiResponse<{
    status: 'online' | 'offline'
    version: string
    services: Record<string, boolean>
  }>> {
    return http.get('/system/status')
  },

  // 获取应用配置
  getConfig(): Promise<ApiResponse<Record<string, any>>> {
    return http.get('/system/config')
  },

  // 检查更新
  checkUpdate(): Promise<ApiResponse<{
    hasUpdate: boolean
    version?: string
    downloadUrl?: string
    changelog?: string
  }>> {
    return http.get('/system/update')
  },

  // 上报错误
  reportError(error: {
    message: string
    stack?: string
    url?: string
    userAgent?: string
    timestamp: number
  }): Promise<ApiResponse<void>> {
    return http.post('/system/error', error, {
      custom: { toast: false } // 错误上报不显示toast
    })
  },

  // 获取帮助文档
  getHelp(): Promise<ApiResponse<Array<{
    id: string
    title: string
    content: string
    category: string
    order: number
  }>>> {
    return http.get('/system/help')
  }
}

// 统计相关API
export const analyticsApi = {
  // 获取使用统计
  getStats(period: 'today' | 'week' | 'month'): Promise<ApiResponse<{
    chatCount: number
    voiceMinutes: number
    messageCount: number
    userActiveTime: number
  }>> {
    return http.get('/analytics/stats', { period })
  },

  // 上报使用事件
  reportEvent(event: {
    type: string
    category: string
    action: string
    value?: number
    properties?: Record<string, any>
  }): Promise<ApiResponse<void>> {
    return http.post('/analytics/event', event, {
      custom: { toast: false, loading: false }
    })
  }
}

// 导出所有API
export const api = {
  auth: authApi,
  user: userApi,
  conversation: conversationApi,
  message: messageApi,
  voice: voiceApi,
  chat: chatApi,
  settings: settingsApi,
  system: systemApi,
  analytics: analyticsApi
}

export default api