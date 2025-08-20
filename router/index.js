import { createRouter, createWebHistory } from "vue-router";

import HomeView from "../src/views/HomeView.vue";

const routes = [
  { path: "/", component: HomeView },
];

export const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
