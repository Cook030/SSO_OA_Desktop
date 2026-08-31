<script setup lang="ts">
/** 下划线风格输入框：小号宽字距标签 + hairline 底线，聚焦时底线加深 */
withDefaults(
  defineProps<{
    label: string;
    type?: string;
    placeholder?: string;
    autocomplete?: string;
    error?: string;
    required?: boolean;
  }>(),
  { type: "text", placeholder: "", autocomplete: "off", error: "", required: false },
);

const model = defineModel<string>({ default: "" });
</script>

<template>
  <label class="block">
    <span class="micro-label">
      {{ label }}
      <span v-if="required" class="text-rust">*</span>
    </span>
    <input
      v-model="model"
      :type="type"
      :placeholder="placeholder"
      :autocomplete="autocomplete"
      class="field-input mt-1"
      :class="{ 'border-rust': error }"
    />
    <span v-if="error" class="mt-1.5 block text-[12px] text-rust">{{ error }}</span>
  </label>
</template>
