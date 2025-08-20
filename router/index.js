import { createRouter, createWebHistory } from "vue-router";

import HomeView from "../src/views/HomeView.vue";
import AdminView from "../src/views/AdminView.vue";

const routes = [
  { path: "/", component: HomeView },
  { path: "/judy", component: AdminView }
];

export const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
