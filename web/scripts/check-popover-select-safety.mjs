import fs from 'node:fs';
import path from 'node:path';

const projectRoot = path.resolve(process.cwd());
const srcRoot = path.join(projectRoot, 'src');
const exts = new Set(['.js', '.jsx', '.ts', '.tsx']);

const triggerViolations = [];
const selectViolations = [];

function walk(dirPath) {
  const entries = fs.readdirSync(dirPath, { withFileTypes: true });
  for (const entry of entries) {
    const fullPath = path.join(dirPath, entry.name);
    if (entry.isDirectory()) {
      walk(fullPath);
      continue;
    }
    if (!exts.has(path.extname(entry.name))) {
      continue;
    }
    inspectFile(fullPath);
  }
}

function getLineNumber(content, index) {
  return content.slice(0, index).split('\n').length;
}

function inspectPopoverBlocks(fullPath, content) {
  const blockRanges = [];
  const tokenPattern = /<AppPopover\b|<\/AppPopover>/g;
  const stack = [];
  let match;
  while ((match = tokenPattern.exec(content)) !== null) {
    if (match[0] === '<AppPopover') {
      stack.push(match.index);
      continue;
    }
    const start = stack.pop();
    if (start !== undefined) {
      blockRanges.push([start, tokenPattern.lastIndex]);
    }
  }

  const relativePath = path.relative(projectRoot, fullPath);
  for (const [start, end] of blockRanges) {
    const block = content.slice(start, end);
    const openingTagEnd = block.indexOf('>');
    const openingTag = openingTagEnd >= 0 ? block.slice(0, openingTagEnd + 1) : block;
    if (/\btrigger=\{/.test(openingTag)) {
      triggerViolations.push(
        `${relativePath}:${getLineNumber(content, start)} AppPopover must use children for the trigger node and string trigger modes like trigger='click'.`,
      );
    }

    const selectPattern = /<AppSelect\b[\s\S]*?\/>/g;
    let selectMatch;
    while ((selectMatch = selectPattern.exec(block)) !== null) {
      const selectTag = selectMatch[0];
      if (/\bgetPopupContainer=/.test(selectTag)) {
        continue;
      }
      const selectStart = start + selectMatch.index;
      selectViolations.push(
        `${relativePath}:${getLineNumber(content, selectStart)} AppSelect inside AppPopover must declare getPopupContainer.`,
      );
    }
  }
}

function inspectFile(fullPath) {
  const content = fs.readFileSync(fullPath, 'utf8');
  if (!content.includes('AppPopover')) {
    return;
  }
  inspectPopoverBlocks(fullPath, content);
}

walk(srcRoot);

if (triggerViolations.length > 0 || selectViolations.length > 0) {
  console.error(
    [
      'Popover interaction safety check failed.',
      'Rules:',
      "1. AppPopover trigger nodes must be passed as children; use trigger='click' for trigger mode.",
      '2. AppSelect rendered inside AppPopover must declare getPopupContainer.',
      '',
      ...triggerViolations.map((item) => `- ${item}`),
      ...selectViolations.map((item) => `- ${item}`),
    ].join('\n'),
  );
  process.exit(1);
}

console.log('popover interaction safety check passed.');
