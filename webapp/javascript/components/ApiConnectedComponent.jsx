import React from 'react';
import "react-dom";
import { buildRenderURL, fetchJSON, fetchNames } from '../util/update_requests';

class ApiConnectedComponent extends React.Component {
  constructor() {
    super()
    
    this.fetchJSON = fetchJSON.bind(this);
    this.fetchNames = fetchNames.bind(this);
    this.buildRenderURL = buildRenderURL.bind(this);
  }

  refreshNames = () => {
    this.fetchNames()
  }

  refreshJson = () => {
    let renderUrl = this.buildRenderURL();
    this.fetchJSON(renderUrl);
  }
}


export default ApiConnectedComponent;