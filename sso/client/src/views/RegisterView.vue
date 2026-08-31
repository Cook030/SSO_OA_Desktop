<script setup lang="ts">
import { reactive, ref } from "vue";
import { useRouter } from "vue-router";

import { ApiError } from "@/api/http";
import { register } from "@/api/auth";
import AuthShell from "@/components/AuthShell.vue";
import TextField from "@/components/TextField.vue";

const router = useRouter();

const form = reactive({
  username: "",
  nickname: "",
  email: "",
  mobile: "",
  password: "",
  confirmPassword: "",
});
const errors = reactive({
  username: "",
  password: "",
  confirmPassword: "",
  email: "",
  form: "",
});
const submitting = ref(false);

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

function validate(): boolean {
  errors.username = form.username.trim() ? "" : "请输入用户名";
  errors.password = form.password.length >= 8 ? "" : "密码至少 8 位";
  errors.confirmPassword =
    form.confirmPassword === form.password ? "" : "两次输入的密码不一致";
  errors.email = !form.email || EMAIL_RE.test(form.email) ? "" : "邮箱格式不正确";
  return !errors.username && !errors.password && !errors.confirmPassword && !errors.email;
}

async function onSubmit(): Promise<void> {
  errors.form = "";
  if (!validate() || submitting.value) return;

  submitting.value = true;
  try {
    await register({
      username: form.username.trim(),
      password: form.password,
      confirmPassword: form.confirmPassword,
      email: form.email.trim() || undefined,
      mobile: form.mobile.trim() || undefined,
      nickname: form.nickname.trim() || undefined,
    });
    await router.replace({ name: "login", query: { registered: "1" } });
  } catch (err) {
    errors.form = err instanceof ApiError ? err.message : "注册失败，请稍后重试";
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <AuthShell image-tag="New · Join MapleHaze">
    <p class="micro-label">Create Account</p>
    <h1 class="mt-4 font-display text-6xl leading-[1.05] font-medium text-ink">Join us.</h1>
    <p class="mt-4 font-display text-lg italic text-ink-soft">
      A single identity, kept calm and portable.
    </p>

    <form class="mt-10 space-y-6" novalidate @submit.prevent="onSubmit">
      <TextField
        v-model="form.username"
        label="用户名 Username"
        placeholder="登录使用的唯一用户名"
        autocomplete="username"
        :error="errors.username"
        required
      />
      <TextField v-model="form.nickname" label="昵称 Nickname" placeholder="可选" />
      <div class="grid grid-cols-1 gap-6 sm:grid-cols-2">
        <TextField
          v-model="form.email"
          label="邮箱 Email"
          type="email"
          placeholder="可选"
          autocomplete="email"
          :error="errors.email"
        />
        <TextField
          v-model="form.mobile"
          label="手机号 Mobile"
          type="tel"
          placeholder="可选"
          autocomplete="tel"
        />
      </div>
      <TextField
        v-model="form.password"
        label="密码 Password"
        type="password"
        placeholder="至少 8 位"
        autocomplete="new-password"
        :error="errors.password"
        required
      />
      <TextField
        v-model="form.confirmPassword"
        label="确认密码 Confirm"
        type="password"
        placeholder="再次输入密码"
        autocomplete="new-password"
        :error="errors.confirmPassword"
        required
      />

      <p v-if="errors.form" class="text-[13px] text-rust">{{ errors.form }}</p>

      <div class="flex items-center justify-between pt-2">
        <button type="submit" class="btn-primary" :disabled="submitting">
          {{ submitting ? "Creating…" : "Create →" }}
        </button>
        <RouterLink to="/login" class="btn-ghost">返回登录</RouterLink>
      </div>
    </form>

    <div class="hairline mt-10 pt-5">
      <p class="text-[13px] text-mist">注册即代表同意内部使用规范与隐私约定。</p>
    </div>
  </AuthShell>
</template>
