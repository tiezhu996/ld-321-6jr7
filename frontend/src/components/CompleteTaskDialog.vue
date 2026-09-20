<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { ElMessage } from 'element-plus';
import type { FormInstance, FormRules } from 'element-plus';
import { completeTask } from '../services/storage.service';
import type { CompleteTaskResult, FarmTask } from '../types/domain';

const props = defineProps<{ visible: boolean; task: FarmTask | null }>();
const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void;
  (e: 'completed', result: CompleteTaskResult): void;
}>();

const dialogVisible = computed({
  get: () => props.visible,
  set: (value: boolean) => emit('update:visible', value),
});

const formRef = ref<FormInstance>();
const submitting = ref(false);
const form = reactive({ actualHours: undefined as number | undefined, fuelLiters: undefined as number | undefined, areaMu: undefined as number | undefined });

// 每次打开对话框时按任务预填：工时默认预计工时，面积默认任务面积，油耗留空必填。
watch(
  () => props.visible,
  (visible) => {
    if (visible && props.task) {
      form.actualHours = props.task.estimatedHours;
      form.fuelLiters = undefined;
      form.areaMu = props.task.areaMu;
      formRef.value?.clearValidate();
    }
  },
);

const rules: FormRules<typeof form> = {
  actualHours: [
    { required: true, message: '请登记实际工时', trigger: 'blur' },
    { type: 'number', min: 0.01, message: '实际工时必须大于 0', trigger: 'blur' },
  ],
  fuelLiters: [
    { required: true, message: '请登记油耗（升）', trigger: 'blur' },
    { type: 'number', min: 0, message: '油耗不能为负数', trigger: 'blur' },
  ],
  areaMu: [
    { required: true, message: '请登记作业面积（亩）', trigger: 'blur' },
    { type: 'number', min: 0.01, message: '作业面积必须大于 0', trigger: 'blur' },
  ],
};

const close = () => emit('update:visible', false);

const handleSubmit = async () => {
  if (!props.task) return;
  const valid = await formRef.value?.validate().catch(() => false);
  if (!valid) return;
  submitting.value = true;
  try {
    const result = await completeTask(props.task.id, {
      actualHours: Number(form.actualHours),
      fuelLiters: Number(form.fuelLiters),
      areaMu: Number(form.areaMu),
    });
    ElMessage.success(`完工登记成功：作业记录 ${result.record.id} 已生成`);
    if (result.maintenanceDue) {
      ElMessage({
        type: 'warning',
        message: `农机 ${result.machineCode} 保养剩余工时已扣减至零，已转为维修中并阻止继续派单`,
        duration: 6000,
      });
    }
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
  <el-dialog v-model="dialogVisible" title="完工登记" width="420px" :close-on-click-modal="false">
    <div v-if="task" class="mb-4 rounded-md bg-slate-50 p-3 text-sm text-slate-600">
      <p><strong>{{ task.type }} · {{ task.field }}</strong></p>
      <p class="mt-1">农机 {{ task.recommendedMachine }} / 驾驶员 {{ task.recommendedDriver }}</p>
    </div>
    <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
      <el-form-item label="实际工时" prop="actualHours">
        <el-input-number v-model="form.actualHours" :min="0" :precision="1" :step="0.5" class="w-full" />
      </el-form-item>
      <el-form-item label="油耗（升）" prop="fuelLiters">
        <el-input-number v-model="form.fuelLiters" :min="0" :precision="1" :step="1" class="w-full" />
      </el-form-item>
      <el-form-item label="作业面积（亩）" prop="areaMu">
        <el-input-number v-model="form.areaMu" :min="0" :precision="1" :step="1" class="w-full" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="close">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">确认完工</el-button>
    </template>
  </el-dialog>
</template>
