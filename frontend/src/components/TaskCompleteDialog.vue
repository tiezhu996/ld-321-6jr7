<script setup lang="ts">
import { reactive, ref, watch } from 'vue';
import { ElMessage } from 'element-plus';
import type { FormInstance, FormRules } from 'element-plus';
import { completeTask } from '../services/storage.service';
import type { CompleteTaskResult, FarmTask } from '../types/domain';

const props = defineProps<{ modelValue: boolean; task: FarmTask | null }>();
const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void;
  (e: 'completed', result: CompleteTaskResult): void;
}>();

const formRef = ref<FormInstance>();
const submitting = ref(false);

const form = reactive({
  actualHours: undefined as number | undefined,
  fuelLiters: undefined as number | undefined,
  areaMu: undefined as number | undefined,
});

const rules: FormRules = {
  actualHours: [{ required: true, type: 'number', min: 0.0001, message: '请输入大于 0 的实际工时', trigger: 'blur' }],
  fuelLiters: [{ required: true, type: 'number', min: 0.0001, message: '请输入大于 0 的实际油耗', trigger: 'blur' }],
  areaMu: [{ required: true, type: 'number', min: 0.0001, message: '请输入大于 0 的作业面积', trigger: 'blur' }],
};

// 每次打开弹窗时重置表单。
watch(
  () => props.modelValue,
  (visible) => {
    if (visible) {
      form.actualHours = undefined;
      form.fuelLiters = undefined;
      form.areaMu = undefined;
      formRef.value?.clearValidate();
    }
  },
);

const close = () => emit('update:modelValue', false);

const handleSubmit = async () => {
  if (!props.task || !formRef.value) {
    return;
  }
  const valid = await formRef.value.validate().catch(() => false);
  if (!valid) {
    return;
  }
  submitting.value = true;
  try {
    const result = await completeTask(props.task.id, {
      actualHours: Number(form.actualHours),
      fuelLiters: Number(form.fuelLiters),
      areaMu: Number(form.areaMu),
    });
    ElMessage.success(result.message);
    emit('completed', result);
    close();
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : '完工登记失败');
  } finally {
    submitting.value = false;
  }
};
</script>

<template>
  <el-dialog
    :model-value="modelValue"
    :title="task ? `完工登记 · ${task.type} · ${task.field}` : '完工登记'"
    width="420px"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <el-form v-if="task" ref="formRef" :model="form" :rules="rules" label-width="110px" @submit.prevent>
      <el-form-item label="作业农机">
        <span class="text-sm text-slate-600">{{ task.recommendedMachine }} / {{ task.recommendedDriver }}</span>
      </el-form-item>
      <el-form-item label="实际工时(h)" prop="actualHours">
        <el-input-number v-model="form.actualHours" :min="0" :precision="1" :step="0.5" class="w-full" />
      </el-form-item>
      <el-form-item label="实际油耗(L)" prop="fuelLiters">
        <el-input-number v-model="form.fuelLiters" :min="0" :precision="1" :step="1" class="w-full" />
      </el-form-item>
      <el-form-item label="作业面积(亩)" prop="areaMu">
        <el-input-number v-model="form.areaMu" :min="0" :precision="1" :step="1" class="w-full" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="close">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">确认完工</el-button>
    </template>
  </el-dialog>
</template>
