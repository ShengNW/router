import React from 'react';
import { Popover } from 'antd';

function AppPopover({ children, content, trigger, ...props }) {
  const triggerNode = children ?? (React.isValidElement(trigger) ? trigger : null);
  const popoverTrigger =
    React.isValidElement(trigger) || trigger === undefined ? undefined : trigger;
  if (!triggerNode) {
    throw new Error(
      'AppPopover requires a trigger node via children. Do not pass JSX to the trigger prop.',
    );
  }
  return (
    <Popover {...props} trigger={popoverTrigger} content={content}>
      {triggerNode}
    </Popover>
  );
}

export default AppPopover;
