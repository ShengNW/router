import React, { useCallback, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Card } from 'semantic-ui-react';
import { API, showError } from '../../helpers';
import { marked } from 'marked';

const About = () => {
  const { t } = useTranslation();
  const [about, setAbout] = useState('');
  const [aboutLoaded, setAboutLoaded] = useState(false);

  const displayAbout = useCallback(async () => {
    setAbout(localStorage.getItem('about') || '');
    const res = await API.get('/api/v1/public/about');
    const { success, message, data } = res.data;
    if (success) {
      let aboutContent = data;
      if (!data.startsWith('https://')) {
        aboutContent = marked.parse(data);
      }
      setAbout(aboutContent);
      localStorage.setItem('about', aboutContent);
    } else {
      showError(message);
      setAbout(t('about.loading_failed'));
    }
    setAboutLoaded(true);
  }, [t]);

  useEffect(() => {
    displayAbout().then();
  }, [displayAbout]);

  return (
    <>
      {aboutLoaded && about === '' ? (
        <div className='dashboard-container'>
          <Card fluid className='chart-card'>
            <Card.Content>
              <Card.Header className='header router-page-title'>{t('about.title')}</Card.Header>
              <p className='router-page-copy'>{t('about.description')}</p>
              {t('about.repository')}
              <a href='https://github.com/yeying-community/router'>
                https://github.com/yeying-community/router
              </a>
            </Card.Content>
          </Card>
        </div>
      ) : (
        <>
          {about.startsWith('https://') ? (
            <iframe
              src={about}
              title='about'
              className='router-embed-frame'
            />
          ) : (
            <div className='dashboard-container'>
              <Card fluid className='chart-card'>
                <Card.Content>
                  <div
                    className='router-content-rendered'
                    dangerouslySetInnerHTML={{ __html: about }}
                  ></div>
                </Card.Content>
              </Card>
            </div>
          )}
        </>
      )}
    </>
  );
};

export default About;
