import React from 'react';
import { Nav } from 'react-bootstrap';

export default class Article extends React.Component {
    render() {
        const servTime = this.props.art.posted_at
        const date = new Date(servTime)
        const postedTime = date.toUTCString()

        return (
            <div className="article">
                <small>Suggested by {this.props.art.user_name} | {postedTime} | from {this.props.art.source_name}</small>
                <h4>{this.props.art.title}</h4>
                <p>{this.props.art.excerpt}</p>
                <Nav.Link className="read-all" href={this.props.art.link}> Read full... </Nav.Link>
                <img src={this.props.art.image_url} alt="" /> 
            </div>
        )
    }
}