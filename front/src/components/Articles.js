import React, { useState, useEffect } from 'react';
import Article from './Article';
import InfiniteScroll from "react-infinite-scroll-component";
import axios from 'axios';
import { toast } from 'react-toastify';

const baseurl = "/home"

const Articles = () => {

    const [currentPage, setCurrentPage] = useState(1)
    const [currentArticles, setCurrentArticles] = useState([])

    useEffect(() => {
        const config = {
            headers: {
                "Authorization": localStorage.getItem('access_token'),
            },
            params: {
                page: currentPage,
            }
        }

        axios.get(baseurl, config).then((res) => {
            if (res.data.body.access_token) {
                localStorage.setItem('access_token', 'Bearer ' + res.data.body.access_token)
            }

            if (res.data.body.user_name) {
                localStorage.setItem('user_name', res.data.body.user_name)
            }

            if (res.data.body.articles) {
                setCurrentArticles(res.data.body.articles)
            }
        })
            .catch((error) => {
                if (error) {
                    toast.error("Failed to load news. Internal server error.")
                    console.error('Internal server error:', error)
                }
            })
    }, [currentPage]);

    const handleNextPage = () => {
        setCurrentPage(currentPage + 1)
    }

    return (
        <>
            {currentArticles.length > 0 ?
                <InfiniteScroll
                    dataLength={currentArticles.length}
                    next={() =>
                        handleNextPage()
                    }
                    hasMore={true}
                >
                    <div className="art">
                        {currentArticles.map((elem) => (
                            <Article art={elem} key={elem.article_id} />
                        ))}
                    </div>
                </InfiniteScroll>
                :
                <div>
                    <h2 className="no-articles">There's nothing to see here yet!</h2>
                </div>}
        </>
    )
}

export default Articles;