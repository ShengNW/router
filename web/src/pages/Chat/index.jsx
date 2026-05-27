import React from 'react';
import { useTranslation } from 'react-i18next';
import { AppFilterHeader } from '../../router-ui';

const Chat = () => {
  const { t } = useTranslation();
  const chatLink = localStorage.getItem('chat_link');

  return (
    <div className='dashboard-container'>
      <AppFilterHeader
        breadcrumbs={[
          { key: 'workspace', label: t('header.user_workspace') },
          { key: 'service', label: t('header.service') },
          { key: 'chat', label: t('header.chat'), active: true },
        ]}
        title={t('header.chat')}
      />
      <iframe
        src={chatLink}
        title={t('header.chat')}
        className='router-embed-frame-chat'
      />
    </div>
  );
};

export default Chat;
