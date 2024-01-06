import React from 'react';
import Artcs from './Artcs';
import AddArtc from './AddArtc';
import axios from 'axios';
import '../css/offer.css';
import { toast } from 'react-toastify';

const articles_url = "/uarticles"
const delete_url = "/adelete"

export default class OfferForm extends React.Component {
    constructor(props) {
        super(props)

        const config = {
            headers: {
                "Authorization": localStorage.getItem('access_token'),
            }
        }

        axios.get(articles_url, config).then((res) => {
            if (res.headers.access_token) {
                localStorage.setItem('access_token', 'Bearer ' + res.headers.access_token)
            }
            this.setState({ articles: res.data })
        })

        this.state = {
            articles: []
        }
        this.deleteArticle = this.deleteArticle.bind(this)
        this.addArticle = this.addArticle.bind(this)
    }
    render() {
        return (<div>
            <h3 className="special-h3">Your unpublished articles</h3>
            <main>
                <Artcs articles={this.state.articles} onDelete={this.deleteArticle} />
            </main>
            <aside>
                <AddArtc onAdd={this.addArticle} />
            </aside>
        </div>)
    }

    deleteArticle(id) {
        const config = {
            headers: {
                "Authorization": localStorage.getItem('access_token'),
                "article_id": id,
            }
        }

        axios.delete(delete_url, config).then((res) => {
            if (res.headers.access_token) {
                localStorage.setItem('access_token', 'Bearer ' + res.headers.access_token)
            }
            if (res.status === 201) {
                toast("Article delited.")
                this.setState({
                    article: this.state.articles.filter((el) => el.id !== id)
                })
            } else if (res.status === 204) {
                toast("Internal server error. Please try again.")
            }
        })
    }

    addArticle(article) {
        const id = article.article_id
        this.setState({ articles: [...this.state.articles, { id, ...article }] })
    }
}