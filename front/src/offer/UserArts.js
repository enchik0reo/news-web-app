import React from "react"
import Art from "./Art"

export default class UserArts extends React.Component {
    render() {
        if (this.props.articles.length > 0)
            return (<div>
                <h3 className="offer-h3">Queue for publication:</h3>
                <small className="offer-small">(Your articles will look like this)</small>
                {this.props.articles.map((elem) => (
                    <Art onEdit={this.props.onEdit} onDelete={this.props.onDelete} key={elem.article_id} art={elem} />
                ))}
            </div>)
        else
            return (<div>
                <h3>Queue for publication is empty</h3>
            </div>)
    }
}
