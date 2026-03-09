import React from 'react';
import { Card } from 'semantic-ui-react';
import UsersTable from '../../components/UsersTable';

const User = () => {
  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <UsersTable />
        </Card.Content>
      </Card>
    </div>
  );
};

export default User;
