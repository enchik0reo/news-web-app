import React from 'react';
import Article from './Article';

export default class Articles extends React.Component {
    render() {
        if(this.props.articles)
            return (<div className="art" >
                {this.props.articles.map((elem) => (
                    <Article art={elem} key={elem.article_id} />
                ))}
            </div>)
        else
            return (<div>
                <h2>There's nothing to see here yet!</h2>
            </div>)
    }
}