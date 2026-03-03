import React from 'react';
import { Outlet } from 'react-router-dom';
import { Container } from 'semantic-ui-react';
import Footer from '../components/Footer';
import Header from '../components/Header';

const AdminLayout = () => (
  <>
    <Header workspace='admin' />
    <Container className='main-content'>
      <Outlet />
    </Container>
    <Footer />
  </>
);

export default AdminLayout;
