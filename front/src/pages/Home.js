import React from 'react';
import '../css/pages.css';
import axios from 'axios';
import Articles from '../components/Articles';

const baseurl = "/home"

export default class Home extends React.Component {
    constructor(props) {
        super(props)

        const config = {
            headers: {
                "Authorization": localStorage.getItem('access_token'),
            }
        }

        axios.get(baseurl, config).then((res) => {
            if (res.headers.access_token) {
                localStorage.setItem('access_token', 'Bearer ' + res.headers.access_token)
            }
            this.setState({ articles: res.data })
        })

        this.state = {
            articles: []
        }
    }

    render() {
        return (
            <>
                <h3 className="special-h3">Latest Articles</h3>
                <Articles articles={this.state.articles} />
            </>
        )
    }
}