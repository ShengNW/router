import {API} from './api';
import {CHANNEL_OPTIONS} from '../constants';

const CHANNEL_TYPES_STORAGE_KEY = 'channel_type_options';

let channelOptions = undefined;
let channelMap = undefined;

const normalizeChannelOption = (item) => {
  if (!item || typeof item !== 'object') {
    return null;
  }
  const rawValue = item.value ?? item.key ?? item.id;
  const value = Number(rawValue);
  if (!Number.isInteger(value) || value < 0) {
    return null;
  }
  const text = String(item.text ?? item.label ?? item.name ?? value).trim();
  return {
    key: value,
    value,
    text: text || String(value),
    color: typeof item.color === 'string' ? item.color.trim() : '',
    description:
      typeof item.description === 'string' ? item.description.trim() : '',
    tip: typeof item.tip === 'string' ? item.tip.trim() : '',
    name: typeof item.name === 'string' ? item.name.trim() : '',
  };
};

const normalizeChannelOptions = (items) => {
  if (!Array.isArray(items)) {
    return [];
  }
  const unique = new Map();
  items.forEach((item) => {
    const normalized = normalizeChannelOption(item);
    if (!normalized) {
      return;
    }
    unique.set(normalized.value, normalized);
  });
  return Array.from(unique.values()).sort((a, b) => a.value - b.value);
};

const updateChannelOptionCache = (items, persist = true) => {
  const normalizedItems = normalizeChannelOptions(items);
  const normalized =
    normalizedItems.length > 0
      ? normalizedItems
      : normalizeChannelOptions(CHANNEL_OPTIONS);
  channelOptions = normalized;
  channelMap = {};
  channelOptions.forEach((option) => {
    channelMap[option.value] = option;
  });
  if (persist && typeof window !== 'undefined') {
    localStorage.setItem(CHANNEL_TYPES_STORAGE_KEY, JSON.stringify(channelOptions));
  }
  return channelOptions;
};

const ensureChannelOptionsLoaded = () => {
  if (channelOptions !== undefined) {
    return channelOptions;
  }
  if (typeof window !== 'undefined') {
    const stored = localStorage.getItem(CHANNEL_TYPES_STORAGE_KEY);
    if (stored) {
      try {
        const parsed = JSON.parse(stored);
        if (Array.isArray(parsed) && parsed.length > 0) {
          return updateChannelOptionCache(parsed, false);
        }
      } catch (e) {
        // ignore invalid cached payload
      }
    }
  }
  return updateChannelOptionCache(CHANNEL_OPTIONS, false);
};

export function getChannelOptions() {
  return ensureChannelOptionsLoaded();
}

export function getChannelOption(channelId) {
  ensureChannelOptionsLoaded();
  return channelMap[channelId];
}

export async function loadChannelOptions() {
  try {
    const res = await API.get('/api/v1/admin/channel/types');
    const {success, data} = res.data || {};
    if (success && Array.isArray(data) && data.length > 0) {
      return updateChannelOptionCache(data, true);
    }
  } catch (e) {
    // keep fallback cache
  }
  return ensureChannelOptionsLoaded();
}
