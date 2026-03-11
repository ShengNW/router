import React, { useState } from 'react';
import { Button, Form, Grid, Header, Image, Message, Card } from 'semantic-ui-react';
import { Link, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { API, getLogo, showError, showInfo, showSuccess } from '../helpers';

const PasswordResetForm = () => {
  const { t } = useTranslation();
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const logo = getLogo();
  const navigate = useNavigate();

  const sendResetEmail = async () => {
    if (email === '') {
      showInfo(t('messages.error.empty_email', '请输入邮箱地址')); 
      return;
    }
    setLoading(true);
    const res = await API.get(`/api/v1/public/reset_password?email=${email}`);
    const { success, message } = res.data;
    setLoading(false);
    if (success) {
      showSuccess(t('messages.success.password_reset'));
      navigate('/login');
    } else {
      showError(message);
    }
  };

  return (
    <Grid textAlign='center' className='router-auth-shell'>
      <Grid.Column>
        <Card fluid className='chart-card router-auth-card'>
          <Card.Content>
            <Card.Header>
              <Header as='h2' textAlign='center' className='router-auth-title router-auth-header'>
                <Image src={logo} className='router-auth-logo' />
                <Header.Content>{t('auth.reset.title')}</Header.Content>
              </Header>
            </Card.Header>
            <Form className='router-auth-form'>
              <Form.Input
                className='router-auth-input'
                fluid
                icon='mail'
                iconPosition='left'
                placeholder={t('auth.reset.email')}
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
              <Button
                className='router-auth-button router-auth-primary'
                fluid
                onClick={sendResetEmail}
                loading={loading}
              >
                {t('auth.reset.button')}
              </Button>
            </Form>

            <Message className='router-auth-message'>
              <div className='router-auth-secondary-text'>
                {t('auth.reset.remember_password')}
                <Link to='/login' className='router-auth-link'>
                  {t('auth.login.login')}
                </Link>
              </div>
            </Message>
          </Card.Content>
        </Card>
      </Grid.Column>
    </Grid>
  );
};

export default PasswordResetForm;
