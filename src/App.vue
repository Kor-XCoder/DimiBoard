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
  opacity: 0.6;
  background-image:
    radial-gradient(4px 4px at 20px 30px, rgba(255, 255, 255, 0.95) 50%, transparent 60%),
    radial-gradient(5px 5px at 120px 80px, rgba(246, 199, 104, 0.9) 50%, transparent 60%),
    radial-gradient(4px 4px at 200px 40px, rgba(125, 175, 255, 0.85) 50%, transparent 60%),
    radial-gradient(6px 6px at 60px 140px, rgba(255, 120, 158, 0.8) 50%, transparent 60%);
  background-size: 220px 220px, 200px 200px, 280px 280px, 260px 260px;
  animation: snow-fall 22s linear infinite, sparkle 6s ease-in-out infinite;
}

@keyframes snow-fall {
  0% {
    background-position: 0 0, 0 0, 0 0, 0 0;
  }
  100% {
    background-position: 0 720px, 0 560px, 0 880px, 0 640px;
  }
}

@keyframes sparkle {
  0%, 100% {
    filter: drop-shadow(0 0 0 rgba(255, 255, 255, 0));
    opacity: 0.55;
  }
  50% {
    filter: drop-shadow(0 0 6px rgba(246, 199, 104, 0.6));
    opacity: 0.75;
  }
}
</style>
