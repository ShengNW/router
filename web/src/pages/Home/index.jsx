import React, { useCallback, useContext, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Card, Grid, Header } from 'semantic-ui-react';
import { API, showError, showNotice, timestamp2string } from '../../helpers';
import { StatusContext } from '../../context/Status';
import { marked } from 'marked';
import { UserContext } from '../../context/User';

const Home = () => {
  const { t } = useTranslation();
  const [statusState] = useContext(StatusContext);
  const [homePageContentLoaded, setHomePageContentLoaded] = useState(false);
  const [homePageContent, setHomePageContent] = useState('');
  const [userState] = useContext(UserContext);

  const displayNotice = useCallback(async () => {
    const res = await API.get('/api/v1/public/notice');
    const { success, message, data } = res.data;
    if (success) {
      let oldNotice = localStorage.getItem('notice');
      if (data !== oldNotice && data !== '') {
        const htmlNotice = marked(data);
        showNotice(htmlNotice, true);
        localStorage.setItem('notice', data);
      }
    } else {
      showError(message);
    }
  }, []);

  const displayHomePageContent = useCallback(async () => {
    setHomePageContent(localStorage.getItem('home_page_content') || '');
    const res = await API.get('/api/v1/public/home_page_content');
    const { success, message, data } = res.data;
    if (success) {
      let content = data;
      if (!data.startsWith('https://')) {
        content = marked.parse(data);
      }
      setHomePageContent(content);
      localStorage.setItem('home_page_content', content);
    } else {
      showError(message);
      setHomePageContent(t('home.loading_failed'));
    }
    setHomePageContentLoaded(true);
  }, [t]);

  const getStartTimeString = () => {
    const timestamp = statusState?.status?.start_time;
    return timestamp2string(timestamp);
  };

  useEffect(() => {
    displayNotice().then();
    displayHomePageContent().then();
  }, [displayNotice, displayHomePageContent]);

  return (
    <>
      {homePageContentLoaded && homePageContent === '' ? (
        <div className='dashboard-container'>
          <Card fluid className='chart-card'>
            <Card.Content>
              <Card.Header className='header router-page-title'>
                {t('home.welcome.title')}
              </Card.Header>
              <Card.Description className='router-page-copy router-copy-relaxed'>
                <p>{t('home.welcome.description')}</p>
                {!userState.user && <p>{t('home.welcome.login_notice')}</p>}
              </Card.Description>
            </Card.Content>
          </Card>
          <Card fluid className='chart-card'>
            <Card.Content>
              <Card.Header>
                <Header as='h3' className='router-section-title'>{t('home.system_status.title')}</Header>
              </Card.Header>
              <Grid columns={2} stackable>
                <Grid.Column>
                  <Card
                    fluid
                    className='chart-card router-soft-card'
                  >
                    <Card.Content>
                      <Card.Header className='router-card-header'>
                        <Header as='h3' className='router-section-title'>
                          {t('home.system_status.info.title')}
                        </Header>
                      </Card.Header>
                      <Card.Description className='router-section-copy router-kv-list router-copy-relaxed'>
                        <p className='router-kv-row'>
                          <i className='info circle icon'></i>
                          <span className='router-kv-key'>
                            {t('home.system_status.info.name')}
                          </span>
                          <span>{statusState?.status?.system_name}</span>
                        </p>
                        <p className='router-kv-row'>
                          <i className='code branch icon'></i>
                          <span className='router-kv-key'>
                            {t('home.system_status.info.version')}
                          </span>
                          <span>
                            {statusState?.status?.version || 'unknown'}
                          </span>
                        </p>
                        <p className='router-kv-row'>
                          <i className='github icon'></i>
                          <span className='router-kv-key'>
                            {t('home.system_status.info.source')}
                          </span>
                          <a
                            href='https://github.com/yeying-community/router'
                            target='_blank'
                            rel='noreferrer'
                            className='router-link-inline'
                          >
                            {t('home.system_status.info.source_link')}
                          </a>
                        </p>
                        <p className='router-kv-row'>
                          <i className='clock outline icon'></i>
                          <span className='router-kv-key'>
                            {t('home.system_status.info.start_time')}
                          </span>
                          <span>{getStartTimeString()}</span>
                        </p>
                      </Card.Description>
                    </Card.Content>
                  </Card>
                </Grid.Column>

                <Grid.Column>
                  <Card
                    fluid
                    className='chart-card router-soft-card'
                  >
                    <Card.Content>
                      <Card.Header className='router-card-header'>
                        <Header as='h3' className='router-section-title'>
                          {t('home.system_status.config.title')}
                        </Header>
                      </Card.Header>
                      <Card.Description className='router-section-copy router-kv-list router-copy-relaxed'>
                        <p className='router-kv-row'>
                          <i className='ethereum icon'></i>
                          <span className='router-kv-key'>
                            {t('home.system_status.config.wallet_login', '钱包登录')}
                          </span>
                          <span
                            className={
                              statusState?.status?.wallet_login
                                ? 'router-kv-value-positive'
                                : 'router-kv-value-negative'
                            }
                          >
                            {statusState?.status?.wallet_login
                              ? t('home.system_status.config.enabled')
                              : t('home.system_status.config.disabled')}
                          </span>
                        </p>
                      </Card.Description>
                    </Card.Content>
                  </Card>
                </Grid.Column>
              </Grid>
            </Card.Content>
          </Card>
        </div>
      ) : (
        <>
          {homePageContent.startsWith('https://') ? (
            <iframe
              src={homePageContent}
              title='home'
              className='router-embed-frame'
            />
          ) : (
            <div
              className='router-page-copy router-content-rendered'
              dangerouslySetInnerHTML={{ __html: homePageContent }}
            ></div>
          )}
        </>
      )}
    </>
  );
};

export default Home;
