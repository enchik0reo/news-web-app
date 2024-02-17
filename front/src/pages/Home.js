import React from 'react';
import Articles from '../components/Articles';
import '../css/pages.css';

export default class Home extends React.Component {
    render() {
        return (
            <>
                <h3 className="special-h3">Latest News</h3>
                <Articles />
            </>
        )
    }
}