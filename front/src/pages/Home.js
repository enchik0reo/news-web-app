import React from 'react';
import '../css/pages.css';
import axios from 'axios';
import Articles from '../components/Articles';
import { toast } from 'react-toastify';

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
            if (res.data.access_token) {
                localStorage.setItem('access_token', 'Bearer ' + res.data.access_token)
            }
        })
            .catch((error) => {
                if (error) {
                    toast.error("Failed to load news. Internal server error.")
                    console.error('Internal server error:', error)
                }
            })

        const ws = new WebSocket('ws://localhost:8008/ws')

        ws.onmessage = (event) => {
            const data = JSON.parse(event.data)
            this.setState({ articles: data.body.articles })
        }

        this.state = {
            articles: []
        }
    }

    render() {
        return (
            <>
                <h3 className="special-h3">Latest News</h3>
                <Articles articles={this.state.articles} />
            </>
        )
    }
}