import React, { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button, Form, Icon, Label, Modal, Segment, Table } from 'semantic-ui-react';
import { API, showError, showInfo, showSuccess, timestamp2string } from '../helpers';

const normalizeProvider = (provider) => {
  if (typeof provider !== 'string') return '';
  const trimmed = provider.trim();
  if (!trimmed) return '';
  const lower = trimmed.toLowerCase();
  switch (lower) {
    case 'gpt':
    case 'openai':
      return 'openai';
    case 'gemini':
    case 'google':
      return 'google';
    case 'claude':
    case 'anthropic':
      return 'anthropic';
    case 'deepseek':
      return 'deepseek';
    case 'qwen':
    case 'qwq':
    case 'qvq':
      return 'qwen';
    default:
      if (trimmed === '千问') return 'qwen';
      return lower;
  }
};

const textToModels = (text) => {
  if (typeof text !== 'string') return [];
  const parts = text
    .split(/\r?\n|,/)
    .map((item) => item.trim())
    .filter((item) => item !== '');
  const seen = new Set();
  const models = [];
  parts.forEach((item) => {
    if (seen.has(item)) return;
    seen.add(item);
    models.push(item);
  });
  return models;
};

const modelsToText = (models) => {
  if (!Array.isArray(models)) return '';
  return models.join('\n');
};

const createEmptyRow = () => ({
  provider: '',
  name: '',
  modelsText: '',
  base_url: '',
  api_key: '',
  source: 'manual',
  updated_at: 0,
});

const toEditableRows = (items) => {
  if (!Array.isArray(items)) return [];
  return items.map((item) => ({
    provider: normalizeProvider(item?.provider || item?.name || ''),
    name: item?.name || '',
    modelsText: modelsToText(item?.models || []),
    base_url: item?.base_url || '',
    api_key: item?.api_key || '',
    source: item?.source || 'manual',
    updated_at: item?.updated_at || 0,
  }));
};

const OFFICIAL_PROVIDER_BASE_URLS = {
  openai: 'https://api.openai.com',
  google: 'https://generativelanguage.googleapis.com/v1beta/openai',
  anthropic: 'https://api.anthropic.com',
  deepseek: 'https://api.deepseek.com',
  qwen: 'https://dashscope.aliyuncs.com/compatible-mode',
};

const ModelProvidersManager = () => {
  const { t } = useTranslation();
  const [rows, setRows] = useState([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [loadingDefaults, setLoadingDefaults] = useState(false);
  const [creating, setCreating] = useState(false);
  const [createRow, setCreateRow] = useState(createEmptyRow());

  const [editing, setEditing] = useState(false);
  const [editIndex, setEditIndex] = useState(-1);
  const [editRow, setEditRow] = useState(createEmptyRow());
  const [fetchingFromApi, setFetchingFromApi] = useState(false);

  const loadCatalog = useCallback(async () => {
    setLoading(true);
    try {
      const res = await API.get('/api/v1/admin/model-provider');
      const { success, message, data } = res.data || {};
      if (!success) {
        showError(message || t('channel.providers.messages.load_failed'));
        return;
      }
      setRows(toEditableRows(data));
    } catch (error) {
      showError(error);
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    loadCatalog().then();
  }, [loadCatalog]);

  const openEditor = (index) => {
    if (index < 0 || index >= rows.length || creating) return;
    setEditIndex(index);
    setEditRow({ ...rows[index] });
    setEditing(true);
  };

  const rollbackEditor = () => {
    setEditing(false);
    setEditIndex(-1);
    setEditRow(createEmptyRow());
    setFetchingFromApi(false);
  };

  const setEditValue = (key, value) => {
    setEditRow((prev) => ({
      ...prev,
      [key]: value,
    }));
  };

  const openCreateModal = () => {
    if (editing) return;
    setCreateRow(createEmptyRow());
    setCreating(true);
  };

  const closeCreateModal = () => {
    setCreating(false);
    setCreateRow(createEmptyRow());
    setFetchingFromApi(false);
  };

  const setCreateValue = (key, value) => {
    setCreateRow((prev) => ({
      ...prev,
      [key]: value,
    }));
  };

  const removeRow = (index) => {
    setRows((prev) => prev.filter((_, idx) => idx !== index));
  };

  const loadDefaults = async () => {
    setLoadingDefaults(true);
    try {
      const res = await API.get('/api/v1/admin/model-provider/defaults');
      const { success, message, data } = res.data || {};
      if (!success) {
        showError(message || t('channel.providers.messages.load_defaults_failed'));
        return;
      }
      setRows(toEditableRows(data));
      showSuccess(t('channel.providers.messages.defaults_loaded'));
    } catch (error) {
      showError(error);
    } finally {
      setLoadingDefaults(false);
    }
  };

  const applyEditToRows = () => {
    const provider = normalizeProvider(editRow.provider);
    if (!provider) {
      showInfo(t('channel.providers.messages.provider_required'));
      return;
    }
    const duplicatedIndex = rows.findIndex(
      (row, index) =>
        index !== editIndex && normalizeProvider(row.provider) === provider
    );
    if (duplicatedIndex !== -1) {
      showInfo(t('channel.providers.messages.provider_exists'));
      return;
    }

    const now = Math.floor(Date.now() / 1000);
    const normalizedRow = {
      ...editRow,
      provider,
      name: (editRow.name || '').trim() || provider,
      modelsText: modelsToText(textToModels(editRow.modelsText)),
      base_url:
        (editRow.base_url || '').trim() ||
        OFFICIAL_PROVIDER_BASE_URLS[provider] ||
        '',
      api_key: (editRow.api_key || '').trim(),
      source: editRow.source || 'manual',
      updated_at: now,
    };

    setRows((prev) => {
      if (editIndex < 0 || editIndex >= prev.length) {
        return [...prev, normalizedRow];
      }
      return prev.map((row, index) => {
        if (index !== editIndex) return row;
        return normalizedRow;
      });
    });
    rollbackEditor();
  };

  const applyCreateToRows = () => {
    const provider = normalizeProvider(createRow.provider);
    if (!provider) {
      showInfo(t('channel.providers.messages.provider_required'));
      return;
    }
    const duplicatedIndex = rows.findIndex(
      (row) => normalizeProvider(row.provider) === provider
    );
    if (duplicatedIndex !== -1) {
      showInfo(t('channel.providers.messages.provider_exists'));
      return;
    }

    const now = Math.floor(Date.now() / 1000);
    const normalizedRow = {
      ...createRow,
      provider,
      name: (createRow.name || '').trim() || provider,
      modelsText: modelsToText(textToModels(createRow.modelsText)),
      base_url:
        (createRow.base_url || '').trim() ||
        OFFICIAL_PROVIDER_BASE_URLS[provider] ||
        '',
      api_key: (createRow.api_key || '').trim(),
      source: createRow.source || 'manual',
      updated_at: now,
    };

    setRows((prev) => [...prev, normalizedRow]);
    closeCreateModal();
  };

  const saveCatalog = async () => {
    const providers = [];
    for (const row of rows) {
      const provider = normalizeProvider(row.provider);
      const name = (row.name || '').trim();
      const models = textToModels(row.modelsText);
      const baseURL = (row.base_url || '').trim();
      const apiKey = (row.api_key || '').trim();
      const hasContent =
        provider || name || models.length > 0 || baseURL !== '' || apiKey !== '';
      if (!hasContent) continue;
      if (!provider) {
        showInfo(t('channel.providers.messages.provider_required'));
        return;
      }
      providers.push({
        provider,
        name: name || provider,
        models,
        base_url: baseURL,
        api_key: apiKey,
        source: row.source || 'manual',
        updated_at: row.updated_at || 0,
      });
    }

    setSaving(true);
    try {
      const res = await API.put('/api/v1/admin/model-provider', {
        providers,
      });
      const { success, message, data } = res.data || {};
      if (!success) {
        showError(message || t('channel.providers.messages.save_failed'));
        return;
      }
      setRows(toEditableRows(data));
      showSuccess(t('channel.providers.messages.save_success'));
    } catch (error) {
      showError(error);
    } finally {
      setSaving(false);
    }
  };

  const fetchModelsFromProviderApi = async (row, onUpdate) => {
    const provider = normalizeProvider(row.provider);
    if (!provider) {
      showInfo(t('channel.providers.messages.fetch_provider_required'));
      return;
    }
    const baseURL =
      (row.base_url || '').trim() || OFFICIAL_PROVIDER_BASE_URLS[provider] || '';
    const apiKey = (row.api_key || '').trim();
    if (!apiKey) {
      showInfo(t('channel.providers.messages.fetch_key_required'));
      return;
    }

    setFetchingFromApi(true);
    try {
      const res = await API.post('/api/v1/admin/model-provider/fetch', {
        provider,
        base_url: baseURL,
        key: apiKey,
      });
      const { success, message, data } = res.data || {};
      if (!success) {
        showError(message || t('channel.providers.messages.fetch_failed'));
        return;
      }

      onUpdate((prev) => ({
        ...prev,
        provider,
        name: (prev.name || '').trim() || provider,
        base_url: baseURL,
        modelsText: modelsToText(Array.isArray(data) ? data : []),
        source: 'api',
        updated_at: Math.floor(Date.now() / 1000),
      }));
      showSuccess(t('channel.providers.messages.fetch_success'));
    } catch (error) {
      showError(error);
    } finally {
      setFetchingFromApi(false);
    }
  };

  const renderRows = () => (
    <Table celled stackable>
      <Table.Header>
        <Table.Row>
          <Table.HeaderCell width={2}>
            {t('channel.providers.table.provider')}
          </Table.HeaderCell>
          <Table.HeaderCell width={2}>
            {t('channel.providers.table.name')}
          </Table.HeaderCell>
          <Table.HeaderCell width={2}>
            {t('channel.providers.table.key')}
          </Table.HeaderCell>
          <Table.HeaderCell width={6}>
            {t('channel.providers.table.models')}
          </Table.HeaderCell>
          <Table.HeaderCell width={1}>
            {t('channel.providers.table.source')}
          </Table.HeaderCell>
          <Table.HeaderCell width={2}>
            {t('channel.providers.table.updated_at')}
          </Table.HeaderCell>
          <Table.HeaderCell width={1}>
            {t('channel.providers.table.actions')}
          </Table.HeaderCell>
        </Table.Row>
      </Table.Header>
      <Table.Body>
        {rows.length === 0 ? (
          <Table.Row>
            <Table.Cell colSpan={7} textAlign='center'>
              {t('channel.providers.table.empty')}
            </Table.Cell>
          </Table.Row>
        ) : (
          rows.map((row, index) => {
            const models = textToModels(row.modelsText);
            const previewModels = models.slice(0, 8);
            const hasMore = models.length > previewModels.length;
            return (
              <Table.Row key={`${row.provider}-${index}`}>
                <Table.Cell>{row.provider || '-'}</Table.Cell>
                <Table.Cell>{row.name || row.provider || '-'}</Table.Cell>
                <Table.Cell textAlign='center'>
                  <Label color={row.api_key ? 'green' : undefined}>
                    {row.api_key
                      ? t('channel.providers.table.key_set')
                      : t('channel.providers.table.key_not_set')}
                  </Label>
                </Table.Cell>
                <Table.Cell>
                  <div style={{ marginBottom: '6px' }}>
                    <Label basic size='tiny'>
                      {t('channel.providers.table.model_count', {
                        count: models.length,
                      })}
                    </Label>
                  </div>
                  <div>
                    {previewModels.map((model) => (
                      <Label
                        key={`${row.provider}-${model}`}
                        size='tiny'
                        style={{ marginBottom: '4px' }}
                      >
                        {model}
                      </Label>
                    ))}
                    {hasMore ? (
                      <Label size='tiny' basic style={{ marginBottom: '4px' }}>
                        +{models.length - previewModels.length}
                      </Label>
                    ) : null}
                  </div>
                </Table.Cell>
                <Table.Cell textAlign='center'>
                  <Label>{row.source || '-'}</Label>
                </Table.Cell>
                <Table.Cell textAlign='center'>
                  {row.updated_at ? timestamp2string(row.updated_at) : '-'}
                </Table.Cell>
                <Table.Cell textAlign='center'>
                  <Button
                    type='button'
                    icon
                    size='tiny'
                    color='blue'
                    disabled={creating}
                    onClick={() => openEditor(index)}
                  >
                    <Icon name='edit' />
                  </Button>
                  <Button
                    type='button'
                    icon
                    size='tiny'
                    color='red'
                    disabled={creating}
                    onClick={() => removeRow(index)}
                  >
                    <Icon name='trash' />
                  </Button>
                </Table.Cell>
              </Table.Row>
            );
          })
        )}
      </Table.Body>
    </Table>
  );

  const renderEditor = () => (
    <Segment>
      <div style={{ fontWeight: 600, marginBottom: 12 }}>
        {editIndex >= 0
          ? t('channel.providers.dialog.title_edit')
          : t('channel.providers.dialog.title_create')}
      </div>
      <Form>
        <Form.Group widths='equal'>
          <Form.Input
            label={t('channel.providers.dialog.provider')}
            placeholder={t('channel.providers.dialog.provider_placeholder')}
            value={editRow.provider}
            onChange={(e, { value }) =>
              setEditValue('provider', normalizeProvider(value || ''))
            }
          />
          <Form.Input
            label={t('channel.providers.dialog.name')}
            placeholder={t('channel.providers.dialog.name_placeholder')}
            value={editRow.name}
            onChange={(e, { value }) => setEditValue('name', value || '')}
          />
        </Form.Group>
        <Form.Group widths='equal'>
          <Form.Input
            label={t('channel.providers.dialog.base_url')}
            placeholder={t('channel.providers.dialog.base_url_placeholder')}
            value={editRow.base_url}
            onChange={(e, { value }) => setEditValue('base_url', value || '')}
          />
          <Form.Input
            label={t('channel.providers.dialog.key')}
            placeholder={t('channel.providers.dialog.key_placeholder')}
            value={editRow.api_key}
            type='password'
            autoComplete='new-password'
            onChange={(e, { value }) => setEditValue('api_key', value || '')}
          />
        </Form.Group>
        <Form.TextArea
          style={{ minHeight: 180, fontFamily: 'JetBrains Mono, Consolas' }}
          label={t('channel.providers.dialog.models')}
          placeholder={t('channel.providers.dialog.models_placeholder')}
          value={editRow.modelsText}
          onChange={(e, { value }) => setEditValue('modelsText', value || '')}
        />
      </Form>

      <div style={{ marginTop: 12 }}>
        <Button
          type='button'
          color='green'
          loading={fetchingFromApi}
          disabled={fetchingFromApi}
          onClick={() => fetchModelsFromProviderApi(editRow, setEditRow)}
        >
          {t('channel.providers.buttons.fetch_from_api')}
        </Button>
        <Button type='button' onClick={rollbackEditor}>
          <Icon name='undo' />
          {t('channel.providers.dialog.cancel')}
        </Button>
        <Button type='button' color='blue' onClick={applyEditToRows}>
          <Icon name='check' />
          {t('channel.providers.dialog.confirm')}
        </Button>
      </div>
    </Segment>
  );

  const renderCreateModal = () => (
    <Modal
      open={creating}
      onClose={closeCreateModal}
      size='small'
      closeOnDimmerClick={false}
    >
      <Modal.Header>{t('channel.providers.dialog.title_create')}</Modal.Header>
      <Modal.Content>
        <Form>
          <Form.Group widths='equal'>
            <Form.Input
              label={t('channel.providers.dialog.provider')}
              placeholder={t('channel.providers.dialog.provider_placeholder')}
              value={createRow.provider}
              onChange={(e, { value }) =>
                setCreateValue('provider', normalizeProvider(value || ''))
              }
            />
            <Form.Input
              label={t('channel.providers.dialog.name')}
              placeholder={t('channel.providers.dialog.name_placeholder')}
              value={createRow.name}
              onChange={(e, { value }) => setCreateValue('name', value || '')}
            />
          </Form.Group>
          <Form.Group widths='equal'>
            <Form.Input
              label={t('channel.providers.dialog.base_url')}
              placeholder={t('channel.providers.dialog.base_url_placeholder')}
              value={createRow.base_url}
              onChange={(e, { value }) => setCreateValue('base_url', value || '')}
            />
            <Form.Input
              label={t('channel.providers.dialog.key')}
              placeholder={t('channel.providers.dialog.key_placeholder')}
              value={createRow.api_key}
              type='password'
              autoComplete='new-password'
              onChange={(e, { value }) => setCreateValue('api_key', value || '')}
            />
          </Form.Group>
          <Form.TextArea
            style={{ minHeight: 180, fontFamily: 'JetBrains Mono, Consolas' }}
            label={t('channel.providers.dialog.models')}
            placeholder={t('channel.providers.dialog.models_placeholder')}
            value={createRow.modelsText}
            onChange={(e, { value }) => setCreateValue('modelsText', value || '')}
          />
        </Form>
      </Modal.Content>
      <Modal.Actions>
        <Button
          type='button'
          color='green'
          loading={fetchingFromApi}
          disabled={fetchingFromApi}
          onClick={() => fetchModelsFromProviderApi(createRow, setCreateRow)}
        >
          {t('channel.providers.buttons.fetch_from_api')}
        </Button>
        <Button type='button' onClick={closeCreateModal}>
          <Icon name='undo' />
          {t('channel.providers.dialog.cancel_create')}
        </Button>
        <Button type='button' color='blue' onClick={applyCreateToRows}>
          <Icon name='check' />
          {t('channel.providers.dialog.confirm')}
        </Button>
      </Modal.Actions>
    </Modal>
  );

  return (
    <div>
      <div style={{ marginBottom: '12px' }}>
        <Button
          type='button'
          onClick={loadCatalog}
          loading={loading}
          disabled={editing || creating}
        >
          {t('channel.providers.buttons.reload')}
        </Button>
        <Button
          type='button'
          onClick={loadDefaults}
          loading={loadingDefaults}
          disabled={loadingDefaults || editing || creating}
        >
          {t('channel.providers.buttons.load_defaults')}
        </Button>
        <Button type='button' onClick={openCreateModal} disabled={editing || creating}>
          {t('channel.providers.buttons.add_provider')}
        </Button>
        <Button
          type='button'
          color='blue'
          loading={saving}
          disabled={saving || editing || creating}
          onClick={saveCatalog}
        >
          {t('channel.providers.buttons.save')}
        </Button>
      </div>

      {renderCreateModal()}
      {editing ? renderEditor() : renderRows()}
    </div>
  );
};

export default ModelProvidersManager;
