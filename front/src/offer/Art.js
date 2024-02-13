import React from "react"
import { IoTrashOutline, IoConstructOutline } from 'react-icons/io5'
import EditArt from "./EditArt"
import { Nav } from 'react-bootstrap';
import empty from '../img/empty.png';

export default class Art extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            editForm: false
        }
    }

    render() {

        var articleImage = this.props.art.image_url
        if (articleImage === "") {
            articleImage = empty
        }

        return (
            <div className="offer-article">
                <IoTrashOutline onClick={() => this.props.onDelete(this.props.art.article_id)} className="delete-icon" />
                <IoConstructOutline onClick={() => {
                    this.setState({
                        editForm: !this.state.editForm
                    })
                }} className="edit-icon" />
                {this.state.editForm && <EditArt id={this.props.art.article_id} onEdit={this.props.onEdit} />}

                <small className="small-info">Suggested by {this.props.art.user_name}</small>
                <h4>{this.props.art.title}</h4>
                <p>{this.props.art.excerpt}</p>
                <Nav.Link className="read-all" href={this.props.art.link}> Read full... </Nav.Link>
                <img src={articleImage} alt="" />
                <small className="small-info">Publication time | {this.props.art.source_name}</small>
            </div>
        )
    }
}
