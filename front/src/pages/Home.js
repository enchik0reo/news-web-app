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
            if (res.headers.access_token) {
                localStorage.setItem('access_token', 'Bearer ' + res.headers.access_token)
            }
            this.setState({ articles: res.data })
        })
            .catch((error) => {
                if (error) {
                    toast.error("Internal server error. Failed to load news.")
                    console.error('Ошибка при выполнении запроса:', error)
                }
            })

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