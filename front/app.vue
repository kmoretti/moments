<template>
  <NuxtLayout>
    <NuxtPage/>
    <Toaster richColors position="top-center" :style="{ top: toastTop + 'px' }"/>
  </NuxtLayout>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { Toaster } from 'vue-sonner'
import "@fancyapps/ui/dist/fancybox/fancybox.css";
import '~/assets/simple-markdown.scss'

const { width } = useWindowSize()
const toastTop = computed(() => width.value < 640 ? 96 : 120)
</script>

<style>
html,
body,
#__nuxt {
  min-height: 100%;
  background: rgb(245 245 244);
}

html.dark,
html.dark body,
html.dark #__nuxt {
  background: rgb(10 10 10);
}

div.grecaptcha-badge{
  display: none !important;
}
</style>
