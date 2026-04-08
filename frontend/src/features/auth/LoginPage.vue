<template>
  <div class="login-page-root" @mousemove="handleMouseMove">
    <!-- Layer 0: Background Layer -->
    <div class="background-layer">
      <div class="grid-pattern"></div>
      <div class="bg-glow-center"></div>
      
      <!-- Parallax Orbs with Fluid Motion -->
      <div class="orb orb-1 orb-animate-1" :style="parallaxStyle(0.05)"></div>
      <div class="orb orb-2 orb-animate-2" :style="parallaxStyle(-0.03)"></div>
      <div class="orb orb-3 orb-animate-3" :style="parallaxStyle(0.02)"></div>

      <!-- Floating Particles -->
      <div class="particles-container">
        <div v-for="i in 20" :key="i" class="particle" :style="randomParticleStyle()"></div>
      </div>
    </div>

    <!-- Layer 1: Main Content Layer -->
    <div class="main-content">
      <!-- Left Column: Info Panel -->
      <div class="info-panel">
        <div class="version-badge reveal-item" style="animation-delay: 0.1s">
          <span class="badge-dot"></span>
          <span class="badge-text">VERSION 1.0 ALPHA</span>
        </div>
        
        <div class="slogan-group">
          <h2 class="slogan-title reveal-item" style="animation-delay: 0.2s">
            Orchestrate your <br /><span class="gradient-text">AI workforce.</span>
          </h2>
          <p class="slogan-sub reveal-item" style="animation-delay: 0.3s">
            Orchestra 是一个为开发者设计的全功能协作平台，集成多智能体管理、实时终端与项目工作区。
          </p>
        </div>

        <div class="feature-list">
          <div class="feature-item reveal-item" style="animation-delay: 0.4s">
            <div class="feature-icon-wrapper">
              <div class="icon-pulse blue-pulse"></div>
              <div class="feature-icon icon-blue">
                <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" /></svg>
              </div>
            </div>
            <div class="feature-text">
              <h3>多智能体协作</h3>
              <p>在同一个界面管理并编排 Claude, Gemini, Aider 等多个智能体。</p>
            </div>
          </div>

          <div class="feature-item reveal-item" style="animation-delay: 0.5s">
            <div class="feature-icon-wrapper">
              <div class="icon-pulse green-pulse"></div>
              <div class="feature-icon icon-green">
                <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" /></svg>
              </div>
            </div>
            <div class="feature-text">
              <h3>原生终端体验</h3>
              <p>基于 PTY 的实时终端，支持 ANSI 全色彩输出与多标签并行操作。</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Right Column: Login Card -->
      <div class="login-card-container reveal-item" style="animation-delay: 0.6s">
        <div class="login-card">
          <div class="shine-sweep"></div>
          
          <div class="card-header">
            <div class="app-logo">
              <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" /></svg>
            </div>
            <h1 class="app-name">Orchestra</h1>
            <p class="app-subtitle">开始你的交响乐</p>
          </div>

          <form @submit.prevent="handleLogin" class="login-form">
            <div class="input-group">
              <label>用户名</label>
              <div class="input-wrapper">
                <div class="input-icon">
                  <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" /></svg>
                </div>
                <input v-model="username" type="text" placeholder="输入您的用户名" :disabled="isSubmitting" />
              </div>
            </div>

            <div class="input-group">
              <label>密码</label>
              <div class="input-wrapper">
                <div class="input-icon">
                  <svg fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" /></svg>
                </div>
                <input v-model="password" type="password" placeholder="输入您的密码" :disabled="isSubmitting" />
              </div>
            </div>

            <div v-if="error" class="error-message">
              {{ error }}
            </div>

            <button type="submit" class="login-btn" :disabled="isSubmitting || !username || !password">
              <div class="btn-content">
                <span v-if="!isSubmitting">登录</span>
                <span v-else>验证中...</span>
                <svg v-if="!isSubmitting" class="btn-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M13 7l5 5m0 0l-5 5m5-5H6" /></svg>
                <div v-else class="spinner"></div>
              </div>
            </button>
          </form>

          <div class="card-footer-note">
            <p>Private Deployment · Invitation Only</p>
          </div>
        </div>
      </div>
    </div>

    <div class="global-footer-simple">
      <p>© 2026 Orchestra AI. 保留所有权利。</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from './authStore'

const router = useRouter()
const authStore = useAuthStore()

const username = ref('')
const password = ref('')
const error = ref('')
const isSubmitting = ref(false)

// Mouse Parallax Logic
const mouse = reactive({ x: 0, y: 0 })
function handleMouseMove(e: MouseEvent) {
  mouse.x = (e.clientX - window.innerWidth / 2) / 15
  mouse.y = (e.clientY - window.innerHeight / 2) / 15
}

function parallaxStyle(factor: number) {
  return {
    transform: `translate(${mouse.x * factor}px, ${mouse.y * factor}px)`
  }
}

function randomParticleStyle() {
  return {
    left: `${Math.random() * 100}%`,
    top: `${Math.random() * 100}%`,
    width: `${Math.random() * 3 + 2}px`,
    height: `${Math.random() * 3 + 2}px`,
    animationDelay: `${Math.random() * 5}s`,
    animationDuration: `${Math.random() * 10 + 10}s`
  }
}

async function handleLogin() {
  error.value = ''
  isSubmitting.value = true
  const success = await authStore.login(username.value, password.value)
  if (success) {
    router.push('/workspaces')
  } else {
    error.value = '用户名或密码错误'
  }
  isSubmitting.value = false
}
</script>

<style scoped>
.login-page-root {
  min-height: 100vh;
  width: 100vw;
  background-color: #f1f5f9;
  position: relative;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #0f172a;
}

/* Layer 0: Background */
.background-layer {
  position: absolute;
  inset: 0;
  z-index: 0;
}

.grid-pattern {
  position: absolute;
  inset: 0;
  background-image: 
    linear-gradient(to right, rgba(15, 23, 42, 0.12) 1px, transparent 1px),
    linear-gradient(to bottom, rgba(15, 23, 42, 0.12) 1px, transparent 1px);
  background-size: 40px 40px;
}

.bg-glow-center {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 100%;
  height: 100%;
  background: radial-gradient(circle at 50% 50%, rgba(99, 102, 241, 0.12), transparent 70%);
}

.orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.2;
  transition: transform 0.1s ease-out;
  pointer-events: none;
}

.orb-1 { top: -10%; left: -10%; width: 50%; height: 50%; background-color: #6366f1; }
.orb-2 { bottom: -10%; right: -5%; width: 45%; height: 45%; background-color: #10b981; }
.orb-3 { top: 30%; right: 15%; width: 25%; height: 25%; background-color: #8b5cf6; opacity: 0.12; }

.particles-container { position: absolute; inset: 0; pointer-events: none; }
.particle {
  position: absolute; background: white; border-radius: 50%; opacity: 0.5;
  box-shadow: 0 0 10px white; animation: float-up infinite linear;
}

@keyframes float-up {
  from { transform: translateY(0); opacity: 0; }
  20% { opacity: 0.6; }
  80% { opacity: 0.6; }
  to { transform: translateY(-100vh); opacity: 0; }
}

/* Layer 1: Content */
.main-content {
  position: relative;
  z-index: 10;
  width: 100%;
  max-width: 1200px;
  padding: 0 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

/* Entrance Animations */
.reveal-item {
  opacity: 0;
  transform: translateY(20px);
  animation: reveal 0.8s forwards cubic-bezier(0.22, 1, 0.36, 1);
}

@keyframes reveal { to { opacity: 1; transform: translateY(0); } }

/* Info Panel */
.info-panel { flex: 0 1 640px; display: flex; flex-direction: column; gap: 48px; }

.version-badge {
  display: flex; align-items: center; gap: 8px; padding: 6px 14px;
  background: white; border: 1px solid rgba(15, 23, 42, 0.15); border-radius: 100px;
  width: fit-content; box-shadow: 0 4px 15px rgba(0,0,0,0.05);
}
.badge-dot { width: 6px; height: 6px; background: #6366f1; border-radius: 50%; box-shadow: 0 0 10px #6366f1; }
.badge-text { font-size: 10px; font-weight: 900; color: #475569; letter-spacing: 0.15em; }

.slogan-title { font-size: 72px; font-weight: 950; line-height: 1.05; color: #0f172a; letter-spacing: -0.03em; }
.gradient-text {
  background: linear-gradient(to right, #4f46e5, #8b5cf6, #10b981);
  background-size: 200% auto; -webkit-background-clip: text; -webkit-text-fill-color: transparent;
  animation: shine-text 6s linear infinite;
}
@keyframes shine-text { to { background-position: 200% center; } }

.slogan-sub { font-size: 20px; color: #475569; line-height: 1.6; max-width: 520px; font-weight: 500; }

.feature-list { display: flex; flex-direction: column; gap: 32px; }
.feature-item { display: flex; gap: 24px; align-items: center; }

.feature-icon-wrapper { position: relative; }
.feature-icon {
  width: 60px; height: 60px; background: white; border-radius: 18px; 
  display: flex; align-items: center; justify-content: center;
  box-shadow: 0 8px 30px rgba(0,0,0,0.08); border: 1px solid white;
  position: relative; z-index: 2;
}
.feature-icon svg { width: 30px; height: 30px; }
.icon-blue svg { color: #6366f1; }
.icon-green svg { color: #10b981; }

.icon-pulse {
  position: absolute; inset: -6px; border-radius: 22px; opacity: 0;
  animation: pulse-ring 3s infinite; z-index: 1;
}
.blue-pulse { border: 2px solid rgba(99, 102, 241, 0.25); }
.green-pulse { border: 2px solid rgba(16, 185, 129, 0.25); animation-delay: 1.5s; }

@keyframes pulse-ring {
  0% { transform: scale(0.85); opacity: 0; }
  50% { opacity: 0.6; }
  100% { transform: scale(1.2); opacity: 0; }
}

.feature-text h3 { font-size: 19px; font-weight: 800; color: #0f172a; margin-bottom: 4px; }
.feature-text p { font-size: 14px; color: #475569; font-weight: 500; }

/* Login Card */
.login-card-container { flex: 0 0 440px; display: flex; justify-content: flex-end; }
.login-card {
  width: 100%; padding: 48px; background: rgba(255, 255, 255, 0.75);
  backdrop-filter: blur(48px); -webkit-backdrop-filter: blur(48px);
  border-radius: 40px; border: 1px solid rgba(255, 255, 255, 0.8);
  box-shadow: 0 40px 100px -20px rgba(0, 0, 0, 0.12), inset 0 0 0 1px rgba(255,255,255,0.5);
  display: flex; flex-direction: column; gap: 40px; position: relative; overflow: hidden;
}

.shine-sweep {
  position: absolute; top: 0; left: -100%; width: 60%; height: 100%;
  background: linear-gradient(to right, transparent, rgba(255,255,255,0.5), transparent);
  transform: skewX(-20deg); animation: sweep 8s infinite ease-in-out;
}
@keyframes sweep { 0% { left: -120%; } 15% { left: 150%; } 100% { left: 150%; } }

.card-header { display: flex; flex-direction: column; gap: 8px; }
.app-logo {
  width: 60px; height: 60px; background: linear-gradient(135deg, #6366f1, #8b5cf6);
  border-radius: 18px; display: flex; align-items: center; justify-content: center;
  color: white; margin-bottom: 16px; box-shadow: 0 12px 25px rgba(99, 102, 241, 0.3);
}
.app-logo svg { width: 34px; height: 34px; }
.app-name { font-size: 34px; font-weight: 900; color: #0f172a; letter-spacing: -0.02em; }
.app-subtitle { font-size: 16px; color: #475569; font-weight: 600; }

.login-form { display: flex; flex-direction: column; gap: 28px; }
.input-group { display: flex; flex-direction: column; gap: 10px; }
.input-group label { font-size: 11px; font-weight: 900; color: #64748b; text-transform: uppercase; letter-spacing: 0.2em; margin-left: 4px; }
.input-wrapper { position: relative; display: flex; align-items: center; }
.input-icon { position: absolute; left: 18px; color: #94a3b8; transition: color 0.3s; z-index: 5; }
.input-wrapper input:focus + .input-icon, .input-wrapper:focus-within .input-icon { color: #6366f1; }

.input-wrapper input {
  width: 100%; padding: 16px 16px 16px 54px; background: rgba(255, 255, 255, 0.65);
  border: 1px solid rgba(15, 23, 42, 0.12); border-radius: 18px;
  font-size: 15px; color: #0f172a; transition: all 0.4s cubic-bezier(0.23, 1, 0.32, 1);
}
.input-wrapper input:focus { background: white; border-color: #6366f1; box-shadow: 0 0 0 6px rgba(99, 102, 241, 0.08); outline: none; }

.error-message { padding: 14px; background: #fef2f2; border: 1px solid #fee2e2; border-radius: 14px; color: #ef4444; font-size: 13px; font-weight: 700; text-align: center; }

.login-btn {
  height: 56px; width: 100%; padding: 0; background: #4f46e5; color: white; 
  border-radius: 18px; border: none; cursor: pointer; transition: all 0.4s;
  box-shadow: 0 15px 35px -5px rgba(79, 70, 229, 0.4); overflow: hidden;
}
.btn-content { width: 100%; height: 100%; display: flex; align-items: center; justify-content: center; gap: 12px; }
.login-btn span { font-size: 16px; font-weight: 900; }
.login-btn:hover:not(:disabled) { background: #4338ca; transform: translateY(-2px); box-shadow: 0 20px 40px -5px rgba(79, 70, 229, 0.5); }
.login-btn:active:not(:disabled) { transform: translateY(0); }
.login-btn:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-icon { width: 20px; height: 20px; }
.spinner { width: 20px; height: 20px; border: 3px solid rgba(255,255,255,0.3); border-top-color: white; border-radius: 50%; animation: spin 0.8s linear infinite; }

.card-footer-note { padding-top: 28px; border-top: 1px solid rgba(15, 23, 42, 0.08); text-align: center; }
.card-footer-note p { font-size: 11px; font-weight: 900; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.15em; opacity: 0.8; }

.global-footer-simple { position: absolute; bottom: 40px; width: 100%; text-align: center; z-index: 20; }
.global-footer-simple p { font-size: 12px; color: #94a3b8; font-weight: 600; }

@media (max-width: 1024px) {
  .main-content { flex-direction: column; padding: 80px 32px; gap: 80px; }
  .info-panel { flex: 1; align-items: center; text-align: center; }
  .slogan-title { font-size: 52px; }
  .login-card-container { flex: 1; width: 100%; justify-content: center; }
}
</style>
