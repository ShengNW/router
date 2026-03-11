import React from 'react';
import { Segment, Dimmer, Loader } from 'semantic-ui-react';

const Loading = ({ prompt: name = 'page' }) => {
  return (
    <Segment className='router-loading-segment'>
      <Dimmer active inverted>
        <Loader indeterminate>加载{name}中...</Loader>
      </Dimmer>
    </Segment>
  );
};

export default Loading;
