<template>
  <div class="api-keys-section">
    <h3 class="section-title">API 密钥管理</h3>
    <p class="section-desc">为不同 AI 提供商配置 API 密钥。密钥将加密存储。</p>

    <!-- Provider List -->
    <div class="providers-grid">
      <div v-for="provider in providers" :key="provider.id" class="provider-card">
        <div class="provider-header">
          <div class="provider-icon">
            <svg v-if="provider.id === 'anthropic'" viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5">
              <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8z"/>
              <path d="M12 6c-3.31 0-6 2.69-6 6s2.69 6 6 6 6-2.69 6-6-2.69-6-6-6z"/>
            </svg>
            <svg v-else-if="provider.id === 'openai'" viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5">
              <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/>
            </svg>
            <svg v-else viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"/>
            </svg>
          </div>
          <div class="provider-info">
            <span class="provider-name">{{ provider.name }}</span>
            <span v-if="getKeyForProvider(provider.id)" class="key-preview">
              {{ getKeyForProvider(provider.id)?.keyPreview }}
            </span>
            <span v-else class="no-key">未配置</span>
          </div>
        </div>

        <div class="provider-actions">
          <button @click="openModal(provider.id)" class="action-btn primary">
            {{ getKeyForProvider(provider.id) ? '更新' : '添加' }}
          </button>
          <button
            v-if="getKeyForProvider(provider.id)"
            @click="testStoredKey(provider.id)"
            class="action-btn secondary"
            :disabled="testing === provider.id"
          >
            {{ testing === provider.id ? '测试中...' : '测试' }}
          </button>
          <button
            v-if="getKeyForProvider(provider.id)"
            @click="handleDelete(getKeyForProvider(provider.id)!.id)"
            class="action-btn danger"
          >
            删除
          </button>
        </div>
      </div>
    </div>

    <!-- Add/Edit Modal -->
    <div v-if="modalOpen" class="modal-overlay" @click.self="closeModal">
      <div class="modal-content">
        <h4 class="modal-title">
          {{ getKeyForProvider(editingProvider) ? '更新' : '添加' }} {{ getProviderName(editingProvider) }} 密钥
        </h4>
        <div class="form-group">
          <label>API 密钥</label>
          <input
            v-model="newKey"
            type="password"
            class="key-input"
            placeholder="输入 API 密钥"
            @keyup.enter="handleSave"
          />
          <p class="input-hint">密钥将使用 AES-GCM 加密存储</p>
        </div>
        <div class="modal-actions">
          <button @click="testNewKey" class="action-btn secondary" :disabled="!newKey || testing === editingProvider">
            {{ testing === editingProvider ? '测试中...' : '测试有效性' }}
          </button>
          <button @click="closeModal" class="action-btn cancel">取消</button>
          <button @click="handleSave" class="action-btn primary" :disabled="!newKey || saving">
            {{ saving ? '保存中...' : '保存' }}
          </button>
        </div>
        <div v-if="testResult" class="test-result" :class="{ valid: testResult.valid, invalid: !testResult.valid }">
          {{ testResult.message }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAPIKeyStore, type APIKeyProvider, type APIKeyResponse, type APIKeyTestResult } from './apiKeyStore'

const store = useAPIKeyStore()

const providers = [
  { id: 'anthropic' as APIKeyProvider, name: 'Anthropic (Claude)' },
  { id: 'openai' as APIKeyProvider, name: 'OpenAI' },
  { id: 'google' as APIKeyProvider, name: 'Google (Gemini)' },
  { id: 'custom' as APIKeyProvider, name: '自定义提供商' },
]

const modalOpen = ref(false)
const editingProvider = ref<APIKeyProvider>('anthropic')
const newKey = ref('')
const saving = ref(false)
const testing = ref<APIKeyProvider | null>(null)
const testResult = ref<APIKeyTestResult | null>(null)

onMounted(() => {
  store.listKeys()
})

function getKeyForProvider(provider: APIKeyProvider): APIKeyResponse | undefined {
  return store.keys.find(k => k.provider === provider)
}

function getProviderName(provider: APIKeyProvider): string {
  return providers.find(p => p.id === provider)?.name || provider
}

function openModal(provider: APIKeyProvider) {
  editingProvider.value = provider
  newKey.value = ''
  testResult.value = null
  modalOpen.value = true
}

function closeModal() {
  modalOpen.value = false
  newKey.value = ''
  testResult.value = null
}

async function handleSave() {
  if (!newKey.value) return
  saving.value = true
  try {
    const success = await store.saveKey(editingProvider.value, newKey.value)
    if (success) {
      closeModal()
    }
  } finally {
    saving.value = false
  }
}

async function handleDelete(id: string) {
  if (confirm('确定要删除此 API 密钥吗？')) {
    await store.deleteKey(id)
  }
}

async function testStoredKey(provider: APIKeyProvider) {
  testing.value = provider
  testResult.value = null
  try {
    testResult.value = await store.testKey(provider)
  } finally {
    testing.value = null
  }
}

async function testNewKey() {
  if (!newKey.value) return
  testing.value = editingProvider.value
  testResult.value = null
  try {
    testResult.value = await store.testKey(editingProvider.value, newKey.value)
  } finally {
    testing.value = null
  }
}
</script>

<style scoped>
.api-keys-section { max-width: 600px; }

.section-title { font-size: 24px; font-weight: 950; color: #0f172a; margin-bottom: 12px; letter-spacing: -0.02em; }
.section-desc { font-size: 14px; color: #64748b; margin-bottom: 32px; }

.providers-grid { display: flex; flex-direction: column; gap: 16px; }

.provider-card {
  background: white; border-radius: 20px; padding: 20px;
  border: 1px solid #e2e8f0; display: flex; align-items: center; justify-content: space-between;
}

.provider-header { display: flex; align-items: center; gap: 16px; }
.provider-icon { width: 40px; height: 40px; border-radius: 12px; background: #f1f5f9; display: flex; align-items: center; justify-content: center; color: #64748b; }
.provider-info { display: flex; flex-direction: column; gap: 4px; }
.provider-name { font-size: 16px; font-weight: 800; color: #0f172a; }
.key-preview { font-size: 13px; color: #4f46e5; font-weight: 600; }
.no-key { font-size: 13px; color: #94a3b8; }

.provider-actions { display: flex; gap: 8px; }
.action-btn {
  padding: 10px 16px; border-radius: 12px; font-size: 13px; font-weight: 700;
  border: none; cursor: pointer; transition: all 0.2s;
}
.action-btn.primary { background: #4f46e5; color: white; }
.action-btn.primary:hover { background: #4338ca; }
.action-btn.secondary { background: #f1f5f9; color: #64748b; }
.action-btn.secondary:hover { background: #e2e8f0; color: #0f172a; }
.action-btn.danger { background: #fef2f2; color: #ef4444; border: 1px solid #fee2e2; }
.action-btn.danger:hover { background: #fee2e2; }
.action-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.action-btn.cancel { background: transparent; color: #64748b; border: 1px solid #e2e8f0; }

/* Modal */
.modal-overlay {
  position: fixed; top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0, 0, 0, 0.4); backdrop-filter: blur(4px);
  display: flex; align-items: center; justify-content: center; z-index: 100;
}

.modal-content {
  background: white; border-radius: 24px; padding: 32px;
  width: 400px; max-width: 90vw; box-shadow: 0 20px 50px rgba(0,0,0,0.2);
}

.modal-title { font-size: 20px; font-weight: 900; color: #0f172a; margin-bottom: 24px; }

.form-group { margin-bottom: 24px; }
.form-group label { font-size: 11px; font-weight: 900; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.15em; margin-bottom: 8px; display: block; }
.key-input {
  width: 100%; padding: 14px 18px; border-radius: 16px; border: 1px solid #e2e8f0;
  background: #f8fafc; color: #0f172a; font-size: 15px; outline: none;
}
.key-input:focus { border-color: #4f46e5; box-shadow: 0 0 0 4px rgba(79, 70, 229, 0.1); }
.input-hint { font-size: 12px; color: #94a3b8; margin-top: 8px; }

.modal-actions { display: flex; gap: 8px; justify-content: flex-end; margin-bottom: 16px; }

.test-result {
  padding: 12px 16px; border-radius: 12px; font-size: 13px; font-weight: 600;
  text-align: center;
}
.test-result.valid { background: #ecfdf5; color: #059669; border: 1px solid #d1fae5; }
.test-result.invalid { background: #fef2f2; color: #ef4444; border: 1px solid #fee2e2; }
</style>