export function resolvePopupContainer(triggerNode) {
  return triggerNode?.parentElement || document.body;
}
