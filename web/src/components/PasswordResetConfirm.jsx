import React, { useEffect, useState } from 'react';
import {
  Button,
  Form,
  Grid,
  Header,
  Image,
  Card,
  Message,
} from 'semantic-ui-react';
import { useTranslation } from 'react-i18next';
import { API, copy, getLogo, showError, showNotice } from '../helpers';
import { useSearchParams } from 'react-router-dom';

const PasswordResetConfirm = () => {
  const { t } = useTranslation();
  const [inputs, setInputs] = useState({
    email: '',
    token: '',
  });
  const { email, token } = inputs;
  const [loading, setLoading] = useState(false);
  const [disableButton, setDisableButton] = useState(false);
  const [newPassword, setNewPassword] = useState('');
  const logo = getLogo();

  const [countdown, setCountdown] = useState(30);

  const [searchParams] = useSearchParams();
  useEffect(() => {
    let token = searchParams.get('token');
    let email = searchParams.get('email');
    setInputs({
      token,
      email,
    });
  }, [searchParams]);

  useEffect(() => {
    let countdownInterval = null;
    if (disableButton && countdown > 0) {
      countdownInterval = setInterval(() => {
        setCountdown(countdown - 1);
      }, 1000);
    } else if (countdown === 0) {
      setDisableButton(false);
      setCountdown(30);
    }
    return () => clearInterval(countdownInterval);
  }, [disableButton, countdown]);

  async function handleSubmit(e) {
    setDisableButton(true);
    if (!email) return;
    setLoading(true);
    const res = await API.post(`/api/v1/public/user/reset`, {
      email,
      token,
    });
    const { success, message } = res.data;
    if (success) {
      let password = res.data.data;
      setNewPassword(password);
      await copy(password);
      showNotice(t('messages.notice.password_copied', { password }));
    } else {
      showError(message);
    }
    setLoading(false);
  }

  return (
    <Grid textAlign='center' className='router-auth-shell'>
      <Grid.Column>
        <Card fluid className='chart-card router-auth-card'>
          <Card.Content>
            <Card.Header>
              <Header
                as='h2'
                textAlign='center'
                className='router-auth-title router-auth-header'
              >
                <Image src={logo} className='router-auth-logo' />
                <Header.Content>{t('auth.reset.confirm.title')}</Header.Content>
              </Header>
            </Card.Header>
            <Form className='router-auth-form'>
              <Form.Input
                className='router-auth-input'
                fluid
                icon='mail'
                iconPosition='left'
                placeholder={t('auth.reset.email')}
                name='email'
                value={email}
                readOnly
              />
              {newPassword && (
                <Form.Input
                  className='router-auth-input router-auth-input-clickable'
                  fluid
                  icon='lock'
                  iconPosition='left'
                  placeholder={t('auth.reset.confirm.new_password')}
                  name='newPassword'
                  value={newPassword}
                  readOnly
                  onClick={(e) => {
                    e.target.select();
                    navigator.clipboard.writeText(newPassword);
                    showNotice(t('auth.reset.confirm.notice'));
                  }}
                />
              )}
              <Button
                className='router-auth-button router-auth-primary'
                fluid
                onClick={handleSubmit}
                loading={loading}
                disabled={disableButton}
              >
                {disableButton
                  ? t('auth.reset.confirm.button_disabled')
                  : t('auth.reset.confirm.button')}
              </Button>
            </Form>
            {newPassword && (
              <Message className='router-auth-message'>
                <p className='router-auth-secondary-text'>
                  {t('auth.reset.confirm.notice')}
                </p>
              </Message>
            )}
          </Card.Content>
        </Card>
      </Grid.Column>
    </Grid>
  );
};

export default PasswordResetConfirm;
