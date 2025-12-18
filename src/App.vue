<template>
  <div id="app">
    <div class="snow-overlay" aria-hidden="true"></div>
    <RouterView />
  </div>
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router';
</script>

<style>
#app {
  width: 100%;
  height: 100dvh;
}

.snow-overlay {
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 999999;
  opacity: 0.2;
  background-image:
    radial-gradient(2px 2px at 20px 30px, rgba(255, 255, 255, 0.95) 50%, transparent 60%),
    radial-gradient(2px 2px at 120px 80px, rgba(255, 255, 255, 0.9) 50%, transparent 60%),
    radial-gradient(2px 2px at 200px 40px, rgba(255, 255, 255, 0.85) 50%, transparent 60%);
  background-size: 240px 240px, 200px 200px, 280px 280px;
  animation: snow-fall 22s linear infinite;
}

@keyframes snow-fall {
  0% {
    background-position: 0 0, 0 0, 0 0;
  }
  100% {
    background-position: 0 720px, 0 560px, 0 880px;
  }
}
</style>