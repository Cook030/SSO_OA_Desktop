<script setup lang="ts">
/**
 * 认证页布局壳：左侧质感视觉区 + 右侧表单区，还原参考图的编辑排版风格
 * - 顶栏品牌标识贯穿
 * - 底部 hairline 之上的小字注解栏
 */
defineProps<{
  /** 左下角图片标签，仿参考图 "NEW · BARRIER SERIES" */
  imageTag?: string;
}>();
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <!-- 顶栏 -->
    <header class="flex items-center justify-between px-6 py-5 lg:px-12">
      <RouterLink to="/login" class="font-display text-lg font-medium tracking-[0.12em] text-ink">
        MAPLEHAZE&nbsp;/&nbsp;SSO
      </RouterLink>
      <nav class="hidden items-center gap-8 text-[13px] text-ink-soft md:flex">
        <span class="cursor-default">统一认证</span>
        <span class="cursor-default">应用接入</span>
        <span class="cursor-default">安全中心</span>
      </nav>
      <span class="text-[13px] text-ink-soft">v0.1</span>
    </header>

    <div class="hairline mx-6 lg:mx-12" />

    <!-- 主体：左图右文 -->
    <main class="flex flex-1 items-stretch">
      <!-- 视觉区 -->
      <section class="relative hidden w-[52%] overflow-hidden lg:block">
        <img
          src="/hero-still.png"
          alt=""
          class="absolute inset-0 h-full w-full object-cover grayscale-[35%]"
        />
        <div class="absolute inset-0 bg-paper/10 mix-blend-screen" />
        <span
          class="absolute bottom-6 left-6 bg-paper px-3 py-1.5 text-[11px] uppercase tracking-[0.24em] text-ink"
        >
          {{ imageTag ?? "MH · SSO Portal" }}
        </span>
      </section>

      <!-- 内容区 -->
      <section class="flex flex-1 items-center justify-center px-6 py-12 lg:px-16">
        <div class="w-full max-w-md">
          <slot />
        </div>
      </section>
    </main>

    <!-- 底部注解栏 -->
    <div class="hairline mx-6 lg:mx-12" />
    <footer
      class="grid grid-cols-2 gap-4 px-6 py-4 text-[12px] text-mist md:grid-cols-4 lg:px-12"
    >
      <span>虚构示例 · 仅用于展示</span>
      <span>认证令牌经 HttpOnly Cookie 下发</span>
      <span>会话支持刷新与撤销</span>
      <span class="md:text-right">MapleHaze 内部门户</span>
    </footer>
  </div>
</template>
