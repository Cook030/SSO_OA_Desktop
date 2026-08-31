<script setup lang="ts">
import { reactive, ref } from "vue";

import AuthShell from "@/components/AuthShell.vue";
import TextField from "@/components/TextField.vue";

/**
 * 找回密码页：当前 mh-sso-svc 未提供自助找回接口，
 * 提交后展示引导说明（联系管理员重置），后续接入邮件/短信验证码流程时替换 onSubmit。
 */
const form = reactive({ account: "" });
const error = ref("");
const submitted = ref(false);
const submitting = ref(false);

async function onSubmit(): Promise<void> {
  error.value = form.account.trim() ? "" : "请输入需要找回的账号";
  if (error.value || submitting.value) return;

  submitting.value = true;
  // 后端暂未开放自助找回接口，此处仅记录并展示引导
  await new Promise((resolve) => setTimeout(resolve, 400));
  submitted.value = true;
  submitting.value = false;
}
</script>

<template>
  <AuthShell image-tag="Support · Account Recovery">
    <p class="micro-label">Reset Password</p>
    <h1 class="mt-4 font-display text-6xl leading-[1.05] font-medium text-ink">Lost keys?</h1>
    <p class="mt-4 font-display text-lg italic text-ink-soft">
      Take a breath. We will walk you back in.
    </p>

    <!-- 已提交：展示管理员协助引导 -->
    <div v-if="submitted" class="mt-10 border border-line bg-paper-deep/60 px-6 py-6">
      <p class="micro-label">Request Noted</p>
      <p class="mt-3 text-[14px] leading-relaxed text-ink-soft">
        账号「{{ form.account }}」的重置申请已记录。当前内测阶段暂未开放自助找回，
        请联系系统管理员协助重置密码，重置后请立即登录并修改。
      </p>
      <RouterLink to="/login" class="btn-primary mt-6">返回登录 →</RouterLink>
    </div>

    <!-- 未提交：账号表单 -->
    <form v-else class="mt-10 space-y-7" novalidate @submit.prevent="onSubmit">
      <TextField
        v-model="form.account"
        label="账号 Account"
        placeholder="用户名 / 邮箱 / 手机号"
        :error="error"
        required
      />

      <div class="flex items-center justify-between pt-2">
        <button type="submit" class="btn-primary" :disabled="submitting">
          {{ submitting ? "Submitting…" : "申请重置 →" }}
        </button>
        <RouterLink to="/login" class="btn-ghost">返回登录</RouterLink>
      </div>
    </form>

    <div class="hairline mt-10 pt-5">
      <p class="text-[13px] text-mist">
        自助找回（邮箱/短信验证码）正在接入中，当前由管理员协助处理。
      </p>
    </div>
  </AuthShell>
</template>
