import React from 'react';
import {
  AppButton,
  AppField,
  AppFormActions,
  AppInput,
  AppModal,
  AppSelect,
} from '../../../router-ui';

const ChannelAppendProviderModal = ({
  t,
  open,
  onClose,
  appendingProviderModel,
  filterProviderOptionsByQuery,
  providerOptions,
  appendProviderForm,
  setAppendProviderForm,
  providerModelTagOptions,
  handleAppendModelToProvider,
}) => {
  return (
    <AppModal
      size='tiny'
      open={open}
      onClose={onClose}
      closeOnDimmerClick={!appendingProviderModel}
      title={t('channel.edit.model_selector.append_dialog.title')}
      footer={
        <AppFormActions>
          <AppButton
            type='button'
            className='router-modal-button'
            onClick={onClose}
            disabled={appendingProviderModel}
          >
            {t('channel.edit.model_selector.append_dialog.cancel')}
          </AppButton>
          <AppButton
            type='button'
            className='router-modal-button'
            color='blue'
            loading={appendingProviderModel}
            disabled={appendingProviderModel}
            onClick={handleAppendModelToProvider}
          >
            {t('channel.edit.model_selector.append_dialog.confirm')}
          </AppButton>
        </AppFormActions>
      }
    >
      <div>
        <div className='router-block-gap'>
          <AppField
            label={t('channel.edit.model_selector.append_dialog.provider')}
          >
            <AppSelect
              search={filterProviderOptionsByQuery}
              className='router-modal-dropdown'
              placeholder={t(
                'channel.edit.model_selector.append_dialog.provider_placeholder',
              )}
              options={providerOptions}
              value={appendProviderForm.provider}
              noResultsMessage={t('common.no_data')}
              onChange={(e, { value }) =>
                setAppendProviderForm((prev) => ({
                  ...prev,
                  provider: (value || '').toString(),
                }))
              }
            />
          </AppField>
          <AppField label={t('channel.edit.model_selector.append_dialog.model')}>
            <AppInput
              className='router-modal-input'
              value={appendProviderForm.model}
              onChange={(e, { value }) =>
                setAppendProviderForm((prev) => ({
                  ...prev,
                  model: value || '',
                }))
              }
            />
          </AppField>
          <AppField label={t('channel.edit.model_selector.append_dialog.tags')}>
            <AppSelect
              className='router-modal-dropdown'
              multiple
              options={providerModelTagOptions}
              value={appendProviderForm.tags || []}
              onChange={(e, { value }) =>
                setAppendProviderForm((prev) => ({
                  ...prev,
                  tags: Array.isArray(value) ? value : [],
                }))
              }
            />
          </AppField>
        </div>
      </div>
    </AppModal>
  );
};

export default ChannelAppendProviderModal;
