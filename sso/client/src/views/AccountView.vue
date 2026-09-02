<script setup lang="ts">
import { computed, reactive, ref, watch } from "vue";
import { useRouter } from "vue-router";

import { authState, changePassword, signOut, updateProfile } from "@/stores/auth";

const router = useRouter();
const signingOut = ref(false);

const user = computed(() => authState.user);
const groups = computed(() => authState.profile?.groups ?? []);
const roles = computed(() => authState.profile?.roles ?? []);

/** 账户信息展示项 */
const infoItems = computed(() => [
  { label: "账号 Account", value: user.value?.account },
  { label: "昵称 Name", value: user.value?.name },
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

// ---------- 编辑资料 ----------

const editing = ref(false);
const saving = ref(false);
const editError = ref("");

const editForm = reactive({ nickname: "", email: "", mobile: "" });

function openEdit(): void {
  editForm.nickname = user.value?.name ?? "";
  editForm.email = user.value?.email ?? "";
  editForm.mobile = user.value?.phone ?? "";
  editError.value = "";
  editing.value = true;
}

function closeEdit(): void {
  if (saving.value) return;
  editing.value = false;
}

async function onSave(): Promise<void> {
  if (saving.value) return;
  editError.value = "";
  saving.value = true;
  try {
    await updateProfile({
      nickname: editForm.nickname.trim(),
      email: editForm.email.trim(),
      mobile: editForm.mobile.trim(),
    });
    editing.value = false;
  } catch (err) {
    editError.value = err instanceof Error ? err.message : "保存失败，请稍后重试";
  } finally {
    saving.value = false;
  }
}

// ---------- 修改密码 ----------

const pwdVisible = ref(false);
const changingPwd = ref(false);
const pwdError = ref("");

const pwdForm = reactive({ password: "", confirmPassword: "" });

function openChangePwd(): void {
  pwdForm.password = "";
  pwdForm.confirmPassword = "";
  pwdError.value = "";
  pwdVisible.value = true;
}

function closeChangePwd(): void {
  if (changingPwd.value) return;
  pwdVisible.value = false;
}

async function onSubmitPassword(): Promise<void> {
  if (changingPwd.value) return;
  pwdError.value = "";

  if (!pwdForm.password) {
    pwdError.value = "请输入新密码";
    return;
  }
  if (pwdForm.password.length < 6) {
    pwdError.value = "新密码长度至少 6 位";
    return;
  }
  if (pwdForm.password !== pwdForm.confirmPassword) {
    pwdError.value = "两次输入的密码不一致";
    return;
  }

  changingPwd.value = true;
  try {
    await changePassword(pwdForm.password, pwdForm.confirmPassword);
    pwdVisible.value = false;
    // 服务端已撤销全部会话并清除 Cookie，需重新登录
    await router.replace({ name: "login", query: { changed: "1" } });
  } catch (err) {
    pwdError.value = err instanceof Error ? err.message : "修改失败，请稍后重试";
  } finally {
    changingPwd.value = false;
  }
}

// 弹窗打开时锁定背景滚动
watch([editing, pwdVisible], ([editOpen, pwdOpen]) => {
  document.body.style.overflow = editOpen || pwdOpen ? "hidden" : "";
});
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
        <div class="mb-2 flex items-center justify-between">
          <p class="micro-label">Profile</p>
          <button class="btn-ghost" @click="openEdit">编辑资料 Edit</button>
        </div>
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
            定期更换密码可提升账户安全。修改成功后需重新登录。
          </p>
          <button class="btn-ghost mt-4" @click="openChangePwd">修改密码 →</button>
        </div>
      </section>
    </main>

    <!-- 编辑资料弹窗 -->
    <div v-if="editing" class="fixed inset-0 z-50 flex items-center justify-center px-6">
      <div
        class="absolute inset-0 bg-ink/30 backdrop-blur-[2px]"
        @click="closeEdit"
      />
      <div class="relative w-full max-w-md border border-line bg-paper px-8 py-8 shadow-xl">
        <p class="micro-label">Edit profile</p>
        <h2 class="mt-3 font-display text-3xl font-medium text-ink">编辑资料</h2>
        <form class="mt-6 space-y-5" @submit.prevent="onSave">
          <div>
            <label class="micro-label mb-1 block" for="edit-nickname">昵称 Name</label>
            <input
              id="edit-nickname"
              v-model="editForm.nickname"
              class="field-input"
              type="text"
              maxlength="64"
              required
            />
          </div>
          <div>
            <label class="micro-label mb-1 block" for="edit-email">邮箱 Email</label>
            <input
              id="edit-email"
              v-model="editForm.email"
              class="field-input"
              type="email"
              maxlength="128"
              placeholder="未填写"
            />
          </div>
          <div>
            <label class="micro-label mb-1 block" for="edit-mobile">手机号 Mobile</label>
            <input
              id="edit-mobile"
              v-model="editForm.mobile"
              class="field-input"
              type="tel"
              maxlength="32"
              placeholder="未填写"
            />
          </div>
          <p v-if="editError" class="text-[13px] text-red-700">{{ editError }}</p>
          <div class="flex items-center justify-end gap-6 pt-2">
            <button type="button" class="btn-ghost" :disabled="saving" @click="closeEdit">
              取消
            </button>
            <button type="submit" class="btn-primary" :disabled="saving">
              {{ saving ? "保存中…" : "保存" }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- 修改密码弹窗 -->
    <div v-if="pwdVisible" class="fixed inset-0 z-50 flex items-center justify-center px-6">
      <div
        class="absolute inset-0 bg-ink/30 backdrop-blur-[2px]"
        @click="closeChangePwd"
      />
      <div class="relative w-full max-w-md border border-line bg-paper px-8 py-8 shadow-xl">
        <p class="micro-label">Security</p>
        <h2 class="mt-3 font-display text-3xl font-medium text-ink">修改密码</h2>
        <p class="mt-2 text-[13px] leading-relaxed text-mist">
          修改成功后所有会话将失效，需使用新密码重新登录。
        </p>
        <form class="mt-6 space-y-5" @submit.prevent="onSubmitPassword">
          <div>
            <label class="micro-label mb-1 block" for="pwd-new">新密码 New password</label>
            <input
              id="pwd-new"
              v-model="pwdForm.password"
              class="field-input"
              type="password"
              minlength="6"
              autocomplete="new-password"
              placeholder="至少 6 位"
              required
            />
          </div>
          <div>
            <label class="micro-label mb-1 block" for="pwd-confirm">确认新密码 Confirm</label>
            <input
              id="pwd-confirm"
              v-model="pwdForm.confirmPassword"
              class="field-input"
              type="password"
              minlength="6"
              autocomplete="new-password"
              placeholder="再次输入新密码"
              required
            />
          </div>
          <p v-if="pwdError" class="text-[13px] text-red-700">{{ pwdError }}</p>
          <div class="flex items-center justify-end gap-6 pt-2">
            <button type="button" class="btn-ghost" :disabled="changingPwd" @click="closeChangePwd">
              取消
            </button>
            <button type="submit" class="btn-primary" :disabled="changingPwd">
              {{ changingPwd ? "提交中…" : "确认修改" }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <div class="hairline mx-6 lg:mx-12" />
    <footer class="px-6 py-4 text-[12px] text-mist lg:px-12">
      MapleHaze 统一认证 · 虚构示例仅用于展示
    </footer>
  </div>
</template>
