<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { setLocale, getCurrentLocale } from '@/locales'
import { useUiStore } from '@/stores/modules/ui'
import { Languages } from 'lucide-vue-next'

const { locale } = useI18n()
const uiStore = useUiStore()

const isOpen = ref(false)

const locales = [
  { code: 'zh-CN', name: '简体中文', flag: '🇨🇳' },
  { code: 'en-US', name: 'English', flag: '🇺🇸' }
]

const currentLocale = computed(() => {
  return locales.find(l => l.code === getCurrentLocale()) || locales[0]
})

const toggleDropdown = () => {
  isOpen.value = !isOpen.value
}

const selectLocale = (localeCode: string) => {
  setLocale(localeCode)
  // 同步更新 Element Plus locale
  uiStore.setLocale(localeCode)
  isOpen.value = false
}

const handleClickOutside = (event: Event) => {
  const dropdown = document.querySelector('[data-language-switcher]')
  if (dropdown && !dropdown.contains(event.target as Node)) {
    isOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
  <div class="relative" data-language-switcher>
    <button
      @click.stop="toggleDropdown"
      class="p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors cursor-pointer"
      :aria-label="$t('language.title') || 'Switch language'"
    >
      <Languages :size="20" class="text-slate-600 dark:text-slate-400" />
    </button>

    <!-- Dropdown Menu -->
    <transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0 scale-95"
      enter-to-class="opacity-100 scale-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-95"
    >
      <div
        v-if="isOpen"
        @click.stop
        class="absolute right-0 mt-2 w-48 bg-white dark:bg-slate-800 rounded-lg shadow-lg border border-slate-200 dark:border-slate-700 py-2 z-50"
      >
        <button
          v-for="loc in locales"
          :key="loc.code"
          @click="selectLocale(loc.code)"
          class="w-full flex items-center gap-3 px-4 py-2 text-sm transition-colors cursor-pointer text-left"
          :class="loc.code === getCurrentLocale()
            ? 'bg-primary-50 dark:bg-primary-900/20 text-primary-600 dark:text-primary-400'
            : 'text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700'"
        >
          <span class="text-lg">{{ loc.flag }}</span>
          <span class="font-medium">{{ loc.name }}</span>
          <svg
            v-if="loc.code === getCurrentLocale()"
            class="w-4 h-4 ml-auto"
            fill="currentColor"
            viewBox="0 0 20 20"
          >
            <path
              fill-rule="evenodd"
              d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
              clip-rule="evenodd"
            />
          </svg>
        </button>
      </div>
    </transition>
  </div>
</template>
