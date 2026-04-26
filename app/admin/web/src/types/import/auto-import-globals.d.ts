export {}

declare global {
  const computed: typeof import('vue')['computed']
  const defineAsyncComponent: typeof import('vue')['defineAsyncComponent']
  const h: typeof import('vue')['h']
  const nextTick: typeof import('vue')['nextTick']
  const onBeforeMount: typeof import('vue')['onBeforeMount']
  const onBeforeUnmount: typeof import('vue')['onBeforeUnmount']
  const onMounted: typeof import('vue')['onMounted']
  const onUnmounted: typeof import('vue')['onUnmounted']
  const reactive: typeof import('vue')['reactive']
  const readonly: typeof import('vue')['readonly']
  const ref: typeof import('vue')['ref']
  const shallowRef: typeof import('vue')['shallowRef']
  const toRefs: typeof import('vue')['toRefs']
  const useAttrs: typeof import('vue')['useAttrs']
  const useTemplateRef: typeof import('vue')['useTemplateRef']
  const watch: typeof import('vue')['watch']

  const storeToRefs: typeof import('pinia')['storeToRefs']

  const useRoute: typeof import('vue-router')['useRoute']
  const useRouter: typeof import('vue-router')['useRouter']

  const useDateFormat: typeof import('@vueuse/core')['useDateFormat']
  const useNow: typeof import('@vueuse/core')['useNow']
  const useScroll: typeof import('@vueuse/core')['useScroll']
  const useWindowSize: typeof import('@vueuse/core')['useWindowSize']

  const ElLoading: typeof import('element-plus/es')['ElLoading']
  const ElMessage: typeof import('element-plus/es')['ElMessage']

  type Ref<T = any> = import('vue').Ref<T>
  type VNode = import('vue').VNode
}
