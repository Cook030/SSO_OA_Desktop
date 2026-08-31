<script setup lang="ts">
import { computed, ref } from "vue";
import { useRouter } from "vue-router";

import { authState, signOut } from "@/stores/auth";

const router = useRouter();
const signingOut = ref(false);

const user = computed(() => authState.user);
const groups = computed(() => authState.profile?.groups ?? []);
const roles = computed(() => authState.profile?.roles ?? []);

/** 账户信息展示项 */
const infoItems = computed(() => [
  { label: "账号 Account", value: user.value?.account },
  { label: "邮箱 Email", value: user.value?.email || "未填写" },
  { label: "手机号 Mobile", value: user.value?.phone || "未填写" },
  {
    label: "角色 Role",
    value: roles.value.map((r) => r.name).join("、") || user.value?.role || "—",
  },
  {
    label: "用户组 Group",
    value: groups.value.map((g) => g.name).join("、") || "—",
  },
  { label: "注册时间 Joined", value: formatTime(user.value?.createTime) },
]);

function formatTime(iso?: string): string {
  if (!iso) return "—";
  const d = new Date(iso);
  return Number.isNaN(d.getTime()) ? iso : d.toLocaleDateString("zh-CN");
}

async function onSignOut(): Promise<void> {
  if (signingOut.value) return;
  signingOut.value = true;
  try {
    await signOut();
  } finally {
    signingOut.value = false;
    await router.replace({ name: "login" });
  }
}
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <!-- 顶栏 -->
    <header class="flex items-center justify-between px-6 py-5 lg:px-12">
      <span class="font-display text-lg font-medium tracking-[0.12em] text-ink">
        MAPLEHAZE&nbsp;/&nbsp;SSO
      </span>
      <button class="btn-ghost" :disabled="signingOut" @click="onSignOut">
        {{ signingOut ? "退出中…" : "退出登录" }}
      </button>
    </header>

    <div class="hairline mx-6 lg:mx-12" />

    <main class="mx-auto w-full max-w-4xl flex-1 px-6 py-14">
      <p class="micro-label">Account</p>
      <h1 class="mt-4 font-display text-6xl leading-[1.05] font-medium text-ink">
        Hello, {{ user?.name || user?.account }}.
      </h1>
      <p class="mt-4 font-display text-lg italic text-ink-soft">
        Your identity, quietly in order.
      </p>

      <!-- 账户信息 -->
      <section class="mt-12">
        <p class="micro-label mb-2">Profile</p>
        <dl>
          <div
            v-for="item in infoItems"
            :key="item.label"
            class="hairline flex items-baseline justify-between gap-6 py-3.5"
          >
            <dt class="text-[12px] uppercase tracking-[0.2em] text-mist">{{ item.label }}</dt>
            <dd class="text-[15px] text-ink">{{ item.value }}</dd>
          </div>
        </dl>
      </section>

      <!-- 功能入口 -->
      <section class="mt-12 grid gap-4 sm:grid-cols-2">
        <div class="border border-line bg-paper-deep/50 px-6 py-6">
          <p class="micro-label">Sessions</p>
          <h2 class="mt-3 font-display text-2xl font-medium text-ink">会话管理</h2>
          <p class="mt-2 text-[13px] leading-relaxed text-mist">
            查看在线设备、远程踢出会话。能力已在服务端就绪，界面即将上线。
          </p>
          <span class="mt-4 inline-block text-[11px] uppercase tracking-[0.22em] text-mist">
            Coming soon
          </span>
        </div>
        <div class="border border-line bg-paper-deep/50 px-6 py-6">
          <p class="micro-label">Security</p>
          <h2 class="mt-3 font-display text-2xl font-medium text-ink">修改密码</h2>
          <p class="mt-2 text-[13px] leading-relaxed text-mist">
            定期更换密码可提升账户安全。界面即将上线。
          </p>
          <span class="mt-4 inline-block text-[11px] uppercase tracking-[0.22em] text-mist">
            Coming soon
          </span>
        </div>
      </section>
    </main>

    <div class="hairline mx-6 lg:mx-12" />
    <footer class="px-6 py-4 text-[12px] text-mist lg:px-12">
      MapleHaze 统一认证 · 虚构示例仅用于展示
    </footer>
  </div>
</template>
