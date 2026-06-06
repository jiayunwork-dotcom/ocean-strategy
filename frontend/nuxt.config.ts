export default defineNuxtConfig({
  devtools: { enabled: true },
  modules: ['@pinia/nuxt', '@nuxtjs/tailwindcss'],
  runtimeConfig: {
    public: {
      apiBase: process.env.API_BASE_URL || 'http://localhost:8080/api/v1',
      wsBase: process.env.WS_BASE_URL || 'ws://localhost:8080/ws'
    }
  },
  css: ['~/assets/css/main.css'],
  app: {
    head: {
      title: '深海策略 - Ocean Strategy',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: '多人回合制深海资源开采策略游戏' }
      ]
    }
  }
})
