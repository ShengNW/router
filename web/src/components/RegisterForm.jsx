import React, { useState } from 'react';
import { Button, Form, Grid, Header, Image, Message, Card } from 'semantic-ui-react';
import { Link, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { API, getLogo, showError, showInfo, showSuccess } from '../helpers';

const RegisterForm = () => {
  const { t } = useTranslation();
  const [inputs, setInputs] = useState({
    username: '',
    password: '',
    password2: '',
    email: '',
  });
  const { username, password, password2 } = inputs;
  const [loading, setLoading] = useState(false);
  const logo = getLogo();
  let affCode = new URLSearchParams(window.location.search).get('aff');
  if (affCode) {
    localStorage.setItem('aff', affCode);
  }

  const navigate = useNavigate();

  function handleChange(e) {
    const { name, value } = e.target;
    setInputs((prev) => ({ ...prev, [name]: value }));
  }

  async function handleSubmit() {
    if (password.length < 8) {
      showInfo(t('messages.error.password_length'));
      return;
    }
    if (password !== password2) {
      showInfo(t('messages.error.password_mismatch'));
      return;
    }
    if (username && password) {
      setLoading(true);
      if (!affCode) {
        affCode = localStorage.getItem('aff');
      }
      const payload = { ...inputs, aff_code: affCode };
      const res = await API.post('/api/v1/public/user/register', payload);
      const { success, message } = res.data;
      setLoading(false);
      if (success) {
        navigate('/login');
        showSuccess(t('messages.success.register'));
      } else {
        showError(message);
      }
    }
  }

  return (
    <Grid textAlign='center' className='router-auth-shell'>
      <Grid.Column>
        <Card fluid className='chart-card router-auth-card'>
          <Card.Content>
            <Card.Header>
              <Header as='h2' textAlign='center' className='router-auth-title router-auth-header'>
                <Image src={logo} className='router-auth-logo' />
                <Header.Content>{t('auth.register.title')}</Header.Content>
              </Header>
            </Card.Header>
            <Form className='router-auth-form'>
              <Form.Input
                className='router-auth-input'
                fluid
                icon='user'
                iconPosition='left'
                placeholder={t('auth.register.username')}
                onChange={handleChange}
                name='username'
              />
              <Form.Input
                className='router-auth-input'
                fluid
                icon='lock'
                iconPosition='left'
                placeholder={t('auth.register.password')}
                onChange={handleChange}
                name='password'
                type='password'
              />
              <Form.Input
                className='router-auth-input'
                fluid
                icon='lock'
                iconPosition='left'
                placeholder={t('auth.register.confirm_password')}
                onChange={handleChange}
                name='password2'
                type='password'
              />
              <Button
                className='router-auth-button router-auth-primary'
                fluid
                onClick={handleSubmit}
                loading={loading}
              >
                {t('auth.register.button')}
              </Button>
            </Form>

            <Message className='router-auth-message'>
              <div className='router-auth-secondary-text'>
                {t('auth.register.has_account')}
                <Link to='/login' className='router-auth-link'>
                  {t('auth.register.login')}
                </Link>
              </div>
            </Message>
          </Card.Content>
        </Card>
      </Grid.Column>
    </Grid>
  );
};

export default RegisterForm;
