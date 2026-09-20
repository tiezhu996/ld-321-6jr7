import { API_BASE } from '../constants/app.constants';
import { AppException } from '../errors/AppException';
import { logger } from '../logger/logger';
import type { CompleteTaskPayload, CompleteTaskResult, DashboardItem, FarmOverview } from '../types/domain';

export const fetchFarmOverview = async (): Promise<FarmOverview> => {
  const response = await fetch(`${API_BASE}/dashboard/overview`);
  if (!response.ok) {
    logger.error('overview request failed', response.status);
    throw new AppException('OVERVIEW_FAILED', '无法加载农机调度看板数据');
  }
  const body = await response.json();
  // 解包后端统一响应 {code, message, data}
  if (body && typeof body === 'object' && body.code === 0 && body.data) {
    return body.data as FarmOverview;
  }
  return body as FarmOverview;
};

export const dispatchTask = async (taskId: string) => {
  const response = await fetch(`${API_BASE}/tasks/${taskId}/dispatch`, { method: 'POST' });
  if (!response.ok) {
    throw new AppException('DISPATCH_FAILED', '派单失败');
  }
  const body = await response.json();
  // 解包后端统一响应 {code, message, data}
  if (body && typeof body === 'object' && body.code === 0 && body.data) {
    return body.data;
  }
  return body;
};

// 读取统一响应中的业务错误消息。
const readErrorMessage = async (response: Response, fallback: string): Promise<string> => {
  try {
    const body = await response.json();
    if (body && typeof body === 'object' && body.message) {
      return body.message as string;
    }
  } catch (err) {
    logger.warn('parse error response failed', err);
  }
  return fallback;
};

// completeTask 已派单任务完工：提交实际工时、油耗、作业面积。
// 重复/并发完工或状态不允许时后端返回 409，字段缺失返回 400。
export const completeTask = async (taskId: string, payload: CompleteTaskPayload): Promise<CompleteTaskResult> => {
  const response = await fetch(`${API_BASE}/tasks/${taskId}/complete`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  if (!response.ok) {
    const message = await readErrorMessage(
      response,
      response.status === 409 ? '任务已完工或当前状态不允许完工' : '完工登记失败，请检查填写内容',
    );
    throw new AppException(response.status === 409 ? 'COMPLETE_CONFLICT' : 'COMPLETE_INVALID', message);
  }
  const body = await response.json();
  if (body && typeof body === 'object' && body.code === 0 && body.data) {
    return body.data as CompleteTaskResult;
  }
  return body as CompleteTaskResult;
};

export const saveItems = (items: DashboardItem[]) => {
  localStorage.setItem('agridispatch.items', JSON.stringify(items));
};
