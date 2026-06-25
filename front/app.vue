<template>
  <NuxtLayout>
    <NuxtPage/>
    <Toaster richColors position="top-center"/>
  </NuxtLayout>
</template>

<script lang="ts" setup>
import { Toaster } from 'vue-sonner'
import "@fancyapps/ui/dist/fancybox/fancybox.css";
import '~/assets/simple-markdown.scss'
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
