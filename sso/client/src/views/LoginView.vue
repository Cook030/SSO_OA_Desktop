<script setup lang="ts">
import { reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";

import { ApiError } from "@/api/http";
import AuthShell from "@/components/AuthShell.vue";
import TextField from "@/components/TextField.vue";
import { signIn } from "@/stores/auth";

const router = useRouter();
const route = useRoute();

const form = reactive({ account: "", password: "" });
const errors = reactive({ account: "", password: "", form: "" });
const submitting = ref(false);

function validate(): boolean {
  errors.account = form.account.trim() ? "" : "请输入登录账号";
  errors.password = form.password ? "" : "请输入密码";
  return !errors.account && !errors.password;
}

async function onSubmit(): Promise<void> {
  errors.form = "";
  if (!validate() || submitting.value) return;

  submitting.value = true;
  try {
    await signIn(form.account.trim(), form.password);
    const redirect = typeof route.query.redirect === "string" ? route.query.redirect : "/account";
    await router.replace(redirect);
  } catch (err) {
    errors.form = err instanceof ApiError ? err.message : "登录失败，请稍后重试";
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <AuthShell image-tag="MH · SSO Portal">
    <p class="micro-label">Single Sign-On</p>
    <h1 class="mt-4 font-display text-6xl leading-[1.05] font-medium text-ink">Welcome back.</h1>
    <p class="mt-4 font-display text-lg italic text-ink-soft">
      One quiet account for every MapleHaze workspace.
    </p>

    <p
      v-if="route.query.changed"
      class="mt-6 border border-pine/40 bg-pine/10 px-4 py-3 text-[13px] text-pine"
    >
      密码修改成功，请使用新密码重新登录
    </p>

    <form class="mt-10 space-y-7" novalidate @submit.prevent="onSubmit">
      <TextField
        v-model="form.account"
        label="账号 Account"
        placeholder="用户名 / 邮箱 / 手机号"
        autocomplete="username"
        :error="errors.account"
        required
      />
      <TextField
        v-model="form.password"
        label="密码 Password"
        type="password"
        placeholder="请输入密码"
        autocomplete="current-password"
        :error="errors.password"
        required
      />

      <p v-if="errors.form" class="text-[13px] text-rust">{{ errors.form }}</p>

      <div class="flex items-center justify-between pt-2">
        <button type="submit" class="btn-primary" :disabled="submitting">
          {{ submitting ? "Signing in…" : "Sign in →" }}
        </button>
        <RouterLink to="/forgot" class="btn-ghost">忘记密码</RouterLink>
      </div>
    </form>

    <div class="hairline mt-10 pt-5">
      <p class="text-[13px] text-mist">账号由企业管理员统一创建，如需开通请联系管理员。</p>
    </div>
  </AuthShell>
</template>
