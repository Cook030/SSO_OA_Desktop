import { createRouter, createWebHistory } from "vue-router";

import { authState, restoreSession } from "@/stores/auth";
import AccountView from "@/views/AccountView.vue";
import ForgotView from "@/views/ForgotView.vue";
import LoginView from "@/views/LoginView.vue";
import RegisterView from "@/views/RegisterView.vue";

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", redirect: "/login" },
    { path: "/login", name: "login", component: LoginView, meta: { guest: true } },
    { path: "/register", name: "register", component: RegisterView, meta: { guest: true } },
    { path: "/forgot", name: "forgot", component: ForgotView, meta: { guest: true } },
    { path: "/account", name: "account", component: AccountView, meta: { requiresAuth: true } },
    { path: "/:pathMatch(.*)*", redirect: "/login" },
  ],
});

router.beforeEach(async (to) => {
  // 首次导航先探测 Cookie 会话（页面刷新后恢复登录态）
  if (!authState.ready) {
    await restoreSession();
  }

  if (to.meta.requiresAuth && !authState.user) {
    return { name: "login", query: { redirect: to.fullPath } };
  }
  // 已登录用户访问登录/注册页时直接进入账户页
  if (to.meta.guest && authState.user) {
    return { name: "account" };
  }
  return true;
});
