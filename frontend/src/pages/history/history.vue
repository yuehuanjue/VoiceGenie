<template>
  <view class="history-container">
    <!-- 搜索栏 -->
    <view class="search-bar">
      <view class="search-input-wrapper">
        <input
          class="search-input"
          type="text"
          placeholder="搜索对话记录..."
          v-model="searchKeyword"
          @input="handleSearch"
        />
        <text class="search-icon">🔍</text>
      </view>
    </view>

    <!-- 过滤选项 -->
    <view class="filter-tabs">
      <view
        class="filter-tab"
        :class="{ active: activeFilter === 'all' }"
        @tap="setFilter('all')"
      >
        全部
      </view>
      <view
        class="filter-tab"
        :class="{ active: activeFilter === 'today' }"
        @tap="setFilter('today')"
      >
        今天
      </view>
      <view
        class="filter-tab"
        :class="{ active: activeFilter === 'week' }"
        @tap="setFilter('week')"
      >
        本周
      </view>
      <view
        class="filter-tab"
        :class="{ active: activeFilter === 'month' }"
        @tap="setFilter('month')"
      >
        本月
      </view>
    </view>

    <!-- 对话记录列表 -->
    <scroll-view
      class="history-list"
      scroll-y="true"
      @scrolltolower="loadMore"
      :refresher-enabled="true"
      :refresher-triggered="isRefreshing"
      @refresherrefresh="onRefresh"
    >
      <view v-if="filteredConversations.length === 0" class="empty-state">
        <view class="empty-icon">📭</view>
        <view class="empty-text">暂无对话记录</view>
        <view class="empty-desc">开始您的第一次AI语音对话吧！</view>
      </view>

      <view
        v-for="conversation in filteredConversations"
        :key="conversation.id"
        class="conversation-item"
        @tap="openConversation(conversation)"
        @longpress="showOptions(conversation)"
      >
        <view class="conversation-header">
          <view class="conversation-title">{{ conversation.title }}</view>
          <view class="conversation-time">{{ formatTime(conversation.updatedAt) }}</view>
        </view>

        <view class="conversation-preview">
          <text class="preview-text">{{ conversation.lastMessage }}</text>
        </view>

        <view class="conversation-meta">
          <view class="meta-item">
            <text class="meta-icon">💬</text>
            <text class="meta-text">{{ conversation.messageCount }}条消息</text>
          </view>
          <view class="meta-item">
            <text class="meta-icon">⏱️</text>
            <text class="meta-text">{{ formatDuration(conversation.duration) }}</text>
          </view>
        </view>
      </view>

      <!-- 加载更多 -->
      <view v-if="hasMore" class="load-more">
        <view v-if="isLoading" class="loading">
          <text class="loading-text">加载中...</text>
        </view>
        <view v-else class="load-more-btn" @tap="loadMore">
          <text>加载更多</text>
        </view>
      </view>
    </scroll-view>

    <!-- 操作菜单 -->
    <view v-if="showActionSheet" class="action-sheet-mask" @tap="hideOptions">
      <view class="action-sheet" @tap.stop>
        <view class="action-sheet-header">
          <text class="action-sheet-title">{{ selectedConversation?.title }}</text>
        </view>
        <view class="action-sheet-content">
          <view class="action-item" @tap="shareConversation">
            <text class="action-icon">📤</text>
            <text class="action-text">分享对话</text>
          </view>
          <view class="action-item" @tap="exportConversation">
            <text class="action-icon">📄</text>
            <text class="action-text">导出记录</text>
          </view>
          <view class="action-item danger" @tap="deleteConversation">
            <text class="action-icon">🗑️</text>
            <text class="action-text">删除对话</text>
          </view>
        </view>
        <view class="action-cancel" @tap="hideOptions">
          <text>取消</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

interface Conversation {
  id: string
  title: string
  lastMessage: string
  messageCount: number
  duration: number
  createdAt: number
  updatedAt: number
}

const searchKeyword = ref<string>('')
const activeFilter = ref<string>('all')
const conversations = ref<Conversation[]>([])
const isLoading = ref<boolean>(false)
const isRefreshing = ref<boolean>(false)
const hasMore = ref<boolean>(true)
const currentPage = ref<number>(1)
const showActionSheet = ref<boolean>(false)
const selectedConversation = ref<Conversation | null>(null)

const filteredConversations = computed(() => {
  let filtered = conversations.value

  // 按时间过滤
  const now = Date.now()
  const oneDayMs = 24 * 60 * 60 * 1000
  const oneWeekMs = 7 * oneDayMs
  const oneMonthMs = 30 * oneDayMs

  switch (activeFilter.value) {
    case 'today':
      filtered = filtered.filter(conv => now - conv.updatedAt < oneDayMs)
      break
    case 'week':
      filtered = filtered.filter(conv => now - conv.updatedAt < oneWeekMs)
      break
    case 'month':
      filtered = filtered.filter(conv => now - conv.updatedAt < oneMonthMs)
      break
  }

  // 按关键词搜索
  if (searchKeyword.value.trim()) {
    const keyword = searchKeyword.value.toLowerCase()
    filtered = filtered.filter(conv =>
      conv.title.toLowerCase().includes(keyword) ||
      conv.lastMessage.toLowerCase().includes(keyword)
    )
  }

  return filtered.sort((a, b) => b.updatedAt - a.updatedAt)
})

onMounted(() => {
  loadConversations()
})

const loadConversations = async (page: number = 1) => {
  if (isLoading.value) return

  isLoading.value = true

  try {
    // 这里会调用API获取对话记录
    // const response = await api.getConversations({ page, limit: 20 })

    // 模拟数据
    const mockData: Conversation[] = Array.from({ length: 10 }, (_, i) => ({
      id: `conv_${page}_${i}`,
      title: `语音对话 ${(page - 1) * 10 + i + 1}`,
      lastMessage: `这是第 ${(page - 1) * 10 + i + 1} 段对话的最后一条消息...`,
      messageCount: Math.floor(Math.random() * 20) + 5,
      duration: Math.floor(Math.random() * 300) + 60,
      createdAt: Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000,
      updatedAt: Date.now() - Math.random() * 7 * 24 * 60 * 60 * 1000
    }))

    if (page === 1) {
      conversations.value = mockData
    } else {
      conversations.value.push(...mockData)
    }

    hasMore.value = page < 5 // 模拟最多5页
    currentPage.value = page

  } catch (error) {
    console.error('加载对话记录失败:', error)
    uni.showToast({
      title: '加载失败',
      icon: 'none'
    })
  } finally {
    isLoading.value = false
    isRefreshing.value = false
  }
}

const handleSearch = () => {
  // 实时搜索，这里可以防抖处理
}

const setFilter = (filter: string) => {
  activeFilter.value = filter
}

const loadMore = () => {
  if (hasMore.value && !isLoading.value) {
    loadConversations(currentPage.value + 1)
  }
}

const onRefresh = () => {
  isRefreshing.value = true
  currentPage.value = 1
  loadConversations(1)
}

const openConversation = (conversation: Conversation) => {
  uni.navigateTo({
    url: `/pages/chat/chat?conversationId=${conversation.id}`
  })
}

const showOptions = (conversation: Conversation) => {
  selectedConversation.value = conversation
  showActionSheet.value = true
}

const hideOptions = () => {
  showActionSheet.value = false
  selectedConversation.value = null
}

const shareConversation = () => {
  // 分享对话
  uni.showToast({
    title: '分享功能开发中',
    icon: 'none'
  })
  hideOptions()
}

const exportConversation = () => {
  // 导出对话记录
  uni.showToast({
    title: '导出功能开发中',
    icon: 'none'
  })
  hideOptions()
}

const deleteConversation = () => {
  if (!selectedConversation.value) return

  uni.showModal({
    title: '确认删除',
    content: '确定要删除这段对话记录吗？删除后无法恢复。',
    success: (res) => {
      if (res.confirm && selectedConversation.value) {
        // 删除对话
        conversations.value = conversations.value.filter(
          conv => conv.id !== selectedConversation.value!.id
        )
        uni.showToast({
          title: '删除成功',
          icon: 'success'
        })
      }
      hideOptions()
    }
  })
}

const formatTime = (timestamp: number): string => {
  const date = new Date(timestamp)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffDays = Math.floor(diffMs / (24 * 60 * 60 * 1000))

  if (diffDays === 0) {
    // 今天，显示时间
    const hours = date.getHours().toString().padStart(2, '0')
    const minutes = date.getMinutes().toString().padStart(2, '0')
    return `${hours}:${minutes}`
  } else if (diffDays === 1) {
    return '昨天'
  } else if (diffDays < 7) {
    return `${diffDays}天前`
  } else {
    // 显示具体日期
    const month = (date.getMonth() + 1).toString().padStart(2, '0')
    const day = date.getDate().toString().padStart(2, '0')
    return `${month}-${day}`
  }
}

const formatDuration = (seconds: number): string => {
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = seconds % 60
  return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`
}
</script>

<style lang="scss" scoped>
.history-container {
  height: 100vh;
  background-color: #f5f5f5;
  display: flex;
  flex-direction: column;
}

.search-bar {
  padding: 20rpx 30rpx;
  background: white;
  border-bottom: 1rpx solid #eee;

  .search-input-wrapper {
    position: relative;
    background: #f8f9fa;
    border-radius: 25rpx;
    padding: 20rpx 60rpx 20rpx 30rpx;

    .search-input {
      width: 100%;
      font-size: 28rpx;
      color: #333;
      border: none;
      outline: none;
      background: transparent;
    }

    .search-icon {
      position: absolute;
      right: 30rpx;
      top: 50%;
      transform: translateY(-50%);
      font-size: 24rpx;
      color: #999;
    }
  }
}

.filter-tabs {
  display: flex;
  background: white;
  border-bottom: 1rpx solid #eee;

  .filter-tab {
    flex: 1;
    text-align: center;
    padding: 25rpx 0;
    font-size: 28rpx;
    color: #666;
    position: relative;

    &.active {
      color: #007AFF;
      font-weight: bold;

      &::after {
        content: '';
        position: absolute;
        bottom: 0;
        left: 50%;
        transform: translateX(-50%);
        width: 60rpx;
        height: 4rpx;
        background: #007AFF;
        border-radius: 2rpx;
      }
    }
  }
}

.history-list {
  flex: 1;
  padding: 20rpx 0;
}

.empty-state {
  padding: 120rpx 40rpx;
  text-align: center;

  .empty-icon {
    font-size: 120rpx;
    margin-bottom: 30rpx;
  }

  .empty-text {
    font-size: 32rpx;
    color: #666;
    margin-bottom: 15rpx;
  }

  .empty-desc {
    font-size: 26rpx;
    color: #999;
  }
}

.conversation-item {
  background: white;
  margin: 0 30rpx 20rpx;
  padding: 30rpx;
  border-radius: 15rpx;
  box-shadow: 0 4rpx 12rpx rgba(0, 0, 0, 0.05);

  .conversation-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20rpx;

    .conversation-title {
      font-size: 32rpx;
      font-weight: bold;
      color: #333;
      flex: 1;
    }

    .conversation-time {
      font-size: 24rpx;
      color: #999;
    }
  }

  .conversation-preview {
    margin-bottom: 20rpx;

    .preview-text {
      font-size: 28rpx;
      color: #666;
      line-height: 1.4;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
      overflow: hidden;
    }
  }

  .conversation-meta {
    display: flex;
    gap: 30rpx;

    .meta-item {
      display: flex;
      align-items: center;
      gap: 8rpx;

      .meta-icon {
        font-size: 20rpx;
      }

      .meta-text {
        font-size: 24rpx;
        color: #999;
      }
    }
  }
}

.load-more {
  text-align: center;
  padding: 40rpx;

  .loading {
    .loading-text {
      font-size: 28rpx;
      color: #999;
    }
  }

  .load-more-btn {
    font-size: 28rpx;
    color: #007AFF;
  }
}

.action-sheet-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 1000;
  display: flex;
  align-items: flex-end;
}

.action-sheet {
  width: 100%;
  background: white;
  border-radius: 20rpx 20rpx 0 0;
  animation: slideUp 0.3s ease-out;

  .action-sheet-header {
    padding: 30rpx;
    border-bottom: 1rpx solid #eee;

    .action-sheet-title {
      font-size: 32rpx;
      font-weight: bold;
      color: #333;
      text-align: center;
    }
  }

  .action-sheet-content {
    .action-item {
      display: flex;
      align-items: center;
      gap: 20rpx;
      padding: 30rpx;
      border-bottom: 1rpx solid #f5f5f5;

      &.danger {
        .action-text {
          color: #ff4757;
        }
      }

      .action-icon {
        font-size: 28rpx;
      }

      .action-text {
        font-size: 32rpx;
        color: #333;
      }
    }
  }

  .action-cancel {
    padding: 30rpx;
    text-align: center;
    font-size: 32rpx;
    color: #666;
    border-top: 10rpx solid #f5f5f5;
  }
}

@keyframes slideUp {
  from {
    transform: translateY(100%);
  }
  to {
    transform: translateY(0);
  }
}
</style>